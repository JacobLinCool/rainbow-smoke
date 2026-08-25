// Package octree indexes active boundary-edge colors in the 8-bit RGB cube.
package octree

import "image/color"

// EmptyRank marks an octree subtree without active boundary edges.
const EmptyRank = ^uint32(0)

// Match is the nearest active RGB leaf and its canonical boundary-edge rank.
type Match struct {
	Color      uint32
	Rank       uint32
	NodeVisits uint32
}

const (
	depth     = 8
	branching = 8
)

// Index is a fixed-depth, dense RGB octree. Updates and queries are not concurrent.
type Index struct {
	nodes   []uint32
	offsets [depth + 1]int
}

// New allocates an empty depth-8 RGB octree.
func New() *Index {
	index := &Index{}
	levelSize := 1
	total := 0
	for level := 0; level <= depth; level++ {
		index.offsets[level] = total
		total += levelSize
		levelSize *= branching
	}

	index.nodes = make([]uint32, total)
	for node := range index.nodes {
		index.nodes[node] = EmptyRank
	}
	return index
}

// Update replaces the minimum active edge rank for one exact 24-bit RGB leaf.
func (index *Index) Update(rgb uint32, rank uint32) {
	rgb &= 0x00ff_ffff
	var path [depth + 1]int
	path[0] = 0
	withinLevel := 0
	for level := 0; level < depth; level++ {
		bit := uint(depth - 1 - level)
		child := int(((rgb>>(16+bit))&1)<<2 | ((rgb>>(8+bit))&1)<<1 | ((rgb >> bit) & 1))
		withinLevel = withinLevel*branching + child
		path[level+1] = index.offsets[level+1] + withinLevel
	}

	leaf := path[depth]
	if index.nodes[leaf] == rank {
		return
	}
	index.nodes[leaf] = rank

	for level := depth - 1; level >= 0; level-- {
		parent := path[level]
		parentWithinLevel := parent - index.offsets[level]
		firstChild := index.offsets[level+1] + parentWithinLevel*branching
		minimum := EmptyRank
		for child := 0; child < branching; child++ {
			minimum = min(minimum, index.nodes[firstChild+child])
		}
		if index.nodes[parent] == minimum {
			break
		}
		index.nodes[parent] = minimum
	}
}

// Nearest returns the active leaf minimizing (squared RGB distance, edge rank).
func (index *Index) Nearest(target color.NRGBA) (Match, bool) {
	if index.nodes[0] == EmptyRank {
		return Match{}, false
	}

	search := nearestSearch{
		index:    index,
		targetR:  int(target.R),
		targetG:  int(target.G),
		targetB:  int(target.B),
		bestDist: int(^uint(0) >> 1),
		bestRank: EmptyRank,
	}
	search.visit(0, 0, 0, 0, 0, 256)
	return Match{
		Color:      search.bestColor,
		Rank:       search.bestRank,
		NodeVisits: search.visits,
	}, true
}

// MemoryBytes returns the byte size of the dense node array.
func (index *Index) MemoryBytes() uint64 {
	return uint64(len(index.nodes)) * 4
}

type nearestSearch struct {
	index     *Index
	targetR   int
	targetG   int
	targetB   int
	bestDist  int
	bestRank  uint32
	bestColor uint32
	visits    uint32
}

type childCandidate struct {
	node        int
	red         int
	green       int
	blue        int
	lowerBound  int
	minimumRank uint32
}

func (search *nearestSearch) visit(node, level, red, green, blue, size int) {
	rank := search.index.nodes[node]
	if rank == EmptyRank {
		return
	}
	search.visits++

	lowerBound := boxDistance(search.targetR, search.targetG, search.targetB, red, green, blue, size)
	if lowerBound > search.bestDist || lowerBound == search.bestDist && rank >= search.bestRank {
		return
	}

	if level == depth {
		search.bestDist = lowerBound
		search.bestRank = rank
		search.bestColor = uint32(red)<<16 | uint32(green)<<8 | uint32(blue)
		return
	}

	half := size / 2
	withinLevel := node - search.index.offsets[level]
	firstChild := search.index.offsets[level+1] + withinLevel*branching
	var candidates [branching]childCandidate
	count := 0
	for child := 0; child < branching; child++ {
		childNode := firstChild + child
		childRank := search.index.nodes[childNode]
		if childRank == EmptyRank {
			continue
		}

		childRed := red
		childGreen := green
		childBlue := blue
		if child&4 != 0 {
			childRed += half
		}
		if child&2 != 0 {
			childGreen += half
		}
		if child&1 != 0 {
			childBlue += half
		}
		candidate := childCandidate{
			node:        childNode,
			red:         childRed,
			green:       childGreen,
			blue:        childBlue,
			lowerBound:  boxDistance(search.targetR, search.targetG, search.targetB, childRed, childGreen, childBlue, half),
			minimumRank: childRank,
		}

		insert := count
		for insert > 0 && candidateLess(candidate, candidates[insert-1]) {
			candidates[insert] = candidates[insert-1]
			insert--
		}
		candidates[insert] = candidate
		count++
	}

	for candidate := 0; candidate < count; candidate++ {
		next := candidates[candidate]
		if next.lowerBound > search.bestDist || next.lowerBound == search.bestDist && next.minimumRank >= search.bestRank {
			continue
		}
		search.visit(next.node, level+1, next.red, next.green, next.blue, half)
	}
}

func candidateLess(left, right childCandidate) bool {
	if left.lowerBound != right.lowerBound {
		return left.lowerBound < right.lowerBound
	}
	return left.minimumRank < right.minimumRank
}

func boxDistance(targetR, targetG, targetB, red, green, blue, size int) int {
	redDistance := axisDistance(targetR, red, red+size-1)
	greenDistance := axisDistance(targetG, green, green+size-1)
	blueDistance := axisDistance(targetB, blue, blue+size-1)
	return redDistance*redDistance + greenDistance*greenDistance + blueDistance*blueDistance
}

func axisDistance(value, minimum, maximum int) int {
	if value < minimum {
		return minimum - value
	}
	if value > maximum {
		return value - maximum
	}
	return 0
}
