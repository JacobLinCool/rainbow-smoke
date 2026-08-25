// Package smoke renders deterministic Rainbow Smoke through one exact octree engine.
package smoke

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/JacobLinCool/rainbow-smoke-go/internal/octree"
)

const (
	maxPixels            = 256 * 256 * 256
	inactiveHeapPosition = int32(-1)
)

var neighbourDirections = [8][2]int{
	{-1, -1},
	{-1, 0},
	{-1, 1},
	{0, -1},
	{0, 1},
	{1, -1},
	{1, 0},
	{1, 1},
}

// Config defines the image geometry and initial painted pixel.
type Config struct {
	Width   int
	Height  int
	CenterX int
	CenterY int
}

// Stats describes placement work without including palette generation or image encoding.
type Stats struct {
	Pixels                int    `json:"pixels"`
	Queries               int    `json:"queries"`
	OctreeNodeVisits      uint64 `json:"octree_node_visits"`
	BoundaryEdgeMutations uint64 `json:"boundary_edge_mutations"`
	MaxFrontier           int    `json:"max_frontier"`
	IndexMemoryBytes      uint64 `json:"index_memory_bytes"`
}

// Engine owns one single-use exact boundary-color index and renderer state.
type Engine struct {
	config        Config
	image         *image.NRGBA
	painted       []bool
	candidates    []int32
	candidatePos  []int32
	edgeHeapPos   []int32
	buckets       map[uint32]*edgeBucket
	index         *octree.Index
	stats         Stats
	renderStarted bool
}

type edgeBucket struct {
	edges       []uint32
	indexedRank uint32
}

// New validates geometry and allocates an empty exact renderer.
func New(config Config) (*Engine, error) {
	if config.Width < 2 || config.Height < 2 {
		return nil, fmt.Errorf("width and height must both be at least 2, got %dx%d", config.Width, config.Height)
	}
	if config.Width > maxPixels/config.Height {
		return nil, fmt.Errorf("image exceeds the %d-pixel RGB space", maxPixels)
	}
	if config.CenterX < 0 || config.CenterX >= config.Width || config.CenterY < 0 || config.CenterY >= config.Height {
		return nil, fmt.Errorf("center (%d,%d) is outside %dx%d image", config.CenterX, config.CenterY, config.Width, config.Height)
	}

	pixelCount := config.Width * config.Height
	if pixelCount > math.MaxInt32/8 {
		return nil, fmt.Errorf("image has too many boundary edges: %d pixels", pixelCount)
	}
	index := octree.New()
	engine := &Engine{
		config:       config,
		image:        image.NewNRGBA(image.Rect(0, 0, config.Width, config.Height)),
		painted:      make([]bool, pixelCount),
		candidates:   make([]int32, 0, pixelCount),
		candidatePos: make([]int32, pixelCount),
		edgeHeapPos:  make([]int32, pixelCount*8),
		buckets:      make(map[uint32]*edgeBucket),
		index:        index,
		stats: Stats{
			Pixels:           pixelCount,
			IndexMemoryBytes: index.MemoryBytes(),
		},
	}
	for pixel := range engine.candidatePos {
		engine.candidatePos[pixel] = -1
	}
	for edge := range engine.edgeHeapPos {
		engine.edgeHeapPos[edge] = inactiveHeapPosition
	}
	return engine, nil
}

// Render consumes the first Width*Height palette entries and returns a complete NRGBA image.
// Cancellation returns an error and no partial image.
func (engine *Engine) Render(ctx context.Context, colors []color.NRGBA) (*image.NRGBA, Stats, error) {
	if engine.renderStarted {
		return nil, engine.stats, fmt.Errorf("engine is single-use")
	}
	engine.renderStarted = true
	pixelCount := engine.config.Width * engine.config.Height
	if len(colors) < pixelCount {
		return nil, engine.stats, fmt.Errorf("palette has %d colors; need %d", len(colors), pixelCount)
	}

	current := int32(engine.config.CenterY*engine.config.Width + engine.config.CenterX)
	for colorIndex := 0; colorIndex < pixelCount; colorIndex++ {
		if colorIndex&1023 == 0 {
			select {
			case <-ctx.Done():
				return nil, engine.stats, ctx.Err()
			default:
			}
		}

		if len(engine.candidates) > 0 {
			match, ok := engine.index.Nearest(colors[colorIndex])
			if !ok {
				return nil, engine.stats, fmt.Errorf("boundary index is empty with %d candidates", len(engine.candidates))
			}
			engine.stats.Queries++
			engine.stats.OctreeNodeVisits += uint64(match.NodeVisits)
			candidateIndex := int(match.Rank >> 3)
			if candidateIndex < 0 || candidateIndex >= len(engine.candidates) {
				return nil, engine.stats, fmt.Errorf("boundary index returned invalid candidate rank %d", match.Rank)
			}
			current = engine.candidates[candidateIndex]
			engine.removeCandidate(candidateIndex)
		}

		engine.setPixel(current, colors[colorIndex])
		engine.painted[current] = true
		engine.updateEdgesToPaintedPixel(current, packColor(colors[colorIndex]))

		var adjacent [8]int32
		count := engine.neighbours(current, &adjacent)
		for index := 0; index < count; index++ {
			pixel := adjacent[index]
			if !engine.painted[pixel] && engine.candidatePos[pixel] < 0 {
				engine.addCandidate(pixel)
			}
		}
		engine.stats.MaxFrontier = max(engine.stats.MaxFrontier, len(engine.candidates))
	}

	return engine.image, engine.stats, nil
}

func (engine *Engine) addCandidate(pixel int32) {
	engine.candidatePos[pixel] = int32(len(engine.candidates))
	engine.candidates = append(engine.candidates, pixel)

	var adjacent [8]int32
	count := engine.neighbours(pixel, &adjacent)
	for slot := 0; slot < count; slot++ {
		engine.pushEdge(engine.edgeID(pixel, slot), engine.pixelColor(adjacent[slot]))
	}
}

func (engine *Engine) removeCandidate(candidateIndex int) {
	pixel := engine.candidates[candidateIndex]
	var adjacent [8]int32
	count := engine.neighbours(pixel, &adjacent)
	for slot := 0; slot < count; slot++ {
		engine.removeEdge(engine.edgeID(pixel, slot), engine.pixelColor(adjacent[slot]))
	}

	lastIndex := len(engine.candidates) - 1
	moved := engine.candidates[lastIndex]
	engine.candidates = engine.candidates[:lastIndex]
	engine.candidatePos[pixel] = -1
	if candidateIndex == lastIndex {
		return
	}

	engine.candidates[candidateIndex] = moved
	engine.candidatePos[moved] = int32(candidateIndex)
	count = engine.neighbours(moved, &adjacent)
	for slot := 0; slot < count; slot++ {
		engine.fixEdge(engine.edgeID(moved, slot), engine.pixelColor(adjacent[slot]))
	}
}

func (engine *Engine) updateEdgesToPaintedPixel(pixel int32, newColor uint32) {
	if newColor == 0 {
		return
	}
	var adjacent [8]int32
	count := engine.neighbours(pixel, &adjacent)
	for index := 0; index < count; index++ {
		candidate := adjacent[index]
		if engine.candidatePos[candidate] < 0 {
			continue
		}
		slot, ok := engine.neighbourSlot(candidate, pixel)
		if !ok {
			panic("smoke: adjacent candidate does not reference painted pixel")
		}
		engine.moveEdge(engine.edgeID(candidate, slot), 0, newColor)
	}
}

func (engine *Engine) pushEdge(edge, colorKey uint32) {
	if engine.edgeHeapPos[edge] != inactiveHeapPosition {
		panic("smoke: boundary edge inserted twice")
	}
	bucket := engine.buckets[colorKey]
	if bucket == nil {
		bucket = &edgeBucket{indexedRank: octree.EmptyRank}
		engine.buckets[colorKey] = bucket
	}
	position := len(bucket.edges)
	bucket.edges = append(bucket.edges, edge)
	engine.edgeHeapPos[edge] = int32(position)
	engine.siftUp(bucket, position)
	engine.syncBucket(colorKey, bucket)
	engine.stats.BoundaryEdgeMutations++
}

func (engine *Engine) removeEdge(edge, colorKey uint32) {
	bucket := engine.buckets[colorKey]
	if bucket == nil {
		panic("smoke: missing boundary color bucket")
	}
	position := int(engine.edgeHeapPos[edge])
	if position < 0 || position >= len(bucket.edges) || bucket.edges[position] != edge {
		panic("smoke: boundary edge heap position is inconsistent")
	}

	last := len(bucket.edges) - 1
	engine.swapEdges(bucket, position, last)
	bucket.edges = bucket.edges[:last]
	engine.edgeHeapPos[edge] = inactiveHeapPosition
	if position < len(bucket.edges) && !engine.siftDown(bucket, position) {
		engine.siftUp(bucket, position)
	}
	engine.syncBucket(colorKey, bucket)
	engine.stats.BoundaryEdgeMutations++
}

func (engine *Engine) moveEdge(edge, oldColor, newColor uint32) {
	if oldColor == newColor {
		return
	}
	engine.removeEdge(edge, oldColor)
	engine.pushEdge(edge, newColor)
}

func (engine *Engine) fixEdge(edge, colorKey uint32) {
	bucket := engine.buckets[colorKey]
	if bucket == nil {
		panic("smoke: missing boundary color bucket during rank update")
	}
	position := int(engine.edgeHeapPos[edge])
	if position < 0 || position >= len(bucket.edges) || bucket.edges[position] != edge {
		panic("smoke: boundary edge rank update has inconsistent heap position")
	}
	if !engine.siftDown(bucket, position) {
		engine.siftUp(bucket, position)
	}
	engine.syncBucket(colorKey, bucket)
}

func (engine *Engine) syncBucket(colorKey uint32, bucket *edgeBucket) {
	minimumRank := octree.EmptyRank
	if len(bucket.edges) > 0 {
		minimumRank = engine.edgeRank(bucket.edges[0])
	}
	if minimumRank != bucket.indexedRank {
		bucket.indexedRank = minimumRank
		engine.index.Update(colorKey, minimumRank)
	}
	if len(bucket.edges) == 0 {
		delete(engine.buckets, colorKey)
	}
}

func (engine *Engine) siftUp(bucket *edgeBucket, position int) {
	for position > 0 {
		parent := (position - 1) / 2
		if !engine.edgeLess(bucket.edges[position], bucket.edges[parent]) {
			return
		}
		engine.swapEdges(bucket, position, parent)
		position = parent
	}
}

func (engine *Engine) siftDown(bucket *edgeBucket, position int) bool {
	start := position
	for {
		left := position*2 + 1
		if left >= len(bucket.edges) {
			break
		}
		smallest := left
		right := left + 1
		if right < len(bucket.edges) && engine.edgeLess(bucket.edges[right], bucket.edges[left]) {
			smallest = right
		}
		if !engine.edgeLess(bucket.edges[smallest], bucket.edges[position]) {
			break
		}
		engine.swapEdges(bucket, position, smallest)
		position = smallest
	}
	return position != start
}

func (engine *Engine) swapEdges(bucket *edgeBucket, left, right int) {
	if left == right {
		return
	}
	bucket.edges[left], bucket.edges[right] = bucket.edges[right], bucket.edges[left]
	engine.edgeHeapPos[bucket.edges[left]] = int32(left)
	engine.edgeHeapPos[bucket.edges[right]] = int32(right)
}

func (engine *Engine) edgeLess(left, right uint32) bool {
	return engine.edgeRank(left) < engine.edgeRank(right)
}

func (engine *Engine) edgeRank(edge uint32) uint32 {
	pixel := edge >> 3
	position := engine.candidatePos[pixel]
	if position < 0 {
		panic("smoke: inactive candidate has an indexed edge")
	}
	return uint32(position)*8 + (edge & 7)
}

func (engine *Engine) edgeID(pixel int32, slot int) uint32 {
	return uint32(pixel)*8 + uint32(slot)
}

func (engine *Engine) neighbours(pixel int32, result *[8]int32) int {
	x := int(pixel) % engine.config.Width
	y := int(pixel) / engine.config.Width
	count := 0
	for _, direction := range neighbourDirections {
		nextX := x + direction[0]
		nextY := y + direction[1]
		if nextX < 0 || nextX >= engine.config.Width || nextY < 0 || nextY >= engine.config.Height {
			continue
		}
		result[count] = int32(nextY*engine.config.Width + nextX)
		count++
	}
	return count
}

func (engine *Engine) neighbourSlot(pixel, target int32) (int, bool) {
	var adjacent [8]int32
	count := engine.neighbours(pixel, &adjacent)
	for slot := 0; slot < count; slot++ {
		if adjacent[slot] == target {
			return slot, true
		}
	}
	return 0, false
}

func (engine *Engine) pixelColor(pixel int32) uint32 {
	if !engine.painted[pixel] {
		return 0
	}
	offset := int(pixel) * 4
	return uint32(engine.image.Pix[offset])<<16 |
		uint32(engine.image.Pix[offset+1])<<8 |
		uint32(engine.image.Pix[offset+2])
}

func (engine *Engine) setPixel(pixel int32, value color.NRGBA) {
	offset := int(pixel) * 4
	engine.image.Pix[offset] = value.R
	engine.image.Pix[offset+1] = value.G
	engine.image.Pix[offset+2] = value.B
	engine.image.Pix[offset+3] = 255
}

func packColor(value color.NRGBA) uint32 {
	return uint32(value.R)<<16 | uint32(value.G)<<8 | uint32(value.B)
}
