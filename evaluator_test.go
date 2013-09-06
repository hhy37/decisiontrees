package decisiontrees

import (
	"code.google.com/p/goprotobuf/proto"
	"flag"
	pb "github.com/ajtulloch/decisiontrees/protobufs"
	"math/rand"
	"sync"
	"testing"
)

var (
	numFeatures = flag.Int("num_features", 1000, "")
	numTrees    = flag.Int("num_trees", 600, "")
	numLevels   = flag.Int("num_levels", 2, "")
)

func makeTree(level int, numFeatures int) *pb.TreeNode {
	if level == 0 {
		return &pb.TreeNode{
			LeafValue: proto.Float64(rand.Float64()),
		}
	}
	splittingFeature := rand.Int63n(int64(numFeatures))
	splittingValue := rand.Float64()
	return &pb.TreeNode{
		Feature:    proto.Int64(splittingFeature),
		SplitValue: proto.Float64(splittingValue),
		Left:       makeTree(level-1, numFeatures),
		Right:      makeTree(level-1, numFeatures),
	}
}

func makeForest(numTrees int, numLevels int, numFeatures int) *pb.Forest {
	forest := &pb.Forest{
		Trees: make([]*pb.TreeNode, numTrees),
	}

	w := sync.WaitGroup{}
	for i := 0; i < numTrees; i++ {
		w.Add(1)
		go func(i int) {
			forest.Trees[i] = makeTree(numLevels, numFeatures)
			w.Done()
		}(i)
	}
	w.Wait()
	return forest
}

func randomFeatureVector(numFeatures int) map[int64]float64 {
	result := make(map[int64]float64)
	for i := 0; i < numFeatures; i++ {
		result[int64(i)] = rand.Float64()
	}
	return result
}

func TestTreeEvaluation(t *testing.T) {
	numFeatures := 1000
	numTrees := 600
	numLevels := 5
	numForests := 1
	numEvaluations := 1
	for i := 0; i < numForests; i++ {
		forest := makeForest(numTrees, numLevels, numFeatures)
		fastEvaluator := NewFastForestEvaluator(forest)
		slowEvaluator := &ForestEvaluator{forest}

		for j := 0; j < numEvaluations; j++ {
			fv := randomFeatureVector(numFeatures)
			fast := fastEvaluator.evaluate(fv)
			slow := slowEvaluator.evaluate(fv)
			if fast != slow {
				t.Errorf("Tree %+v, fast: %v, slow: %v", forest.String(), fast, slow)
			}
		}
	}
}

func benchEvaluator(f func(*pb.Forest) Evaluator, b *testing.B) {
	forest := makeForest(*numTrees, *numLevels, *numFeatures)
	evaluator := f(forest)

	featureVectors := make([]map[int64]float64, 0, b.N)
	for i := 0; i < b.N; i++ {
		featureVectors = append(featureVectors, randomFeatureVector(*numFeatures))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		evaluator.evaluate(featureVectors[i])
	}
}

func BenchmarkFastTreeEvaluation(b *testing.B) {
	flag.Parse()
	f := func(forest *pb.Forest) Evaluator {
		return NewFastForestEvaluator(forest)
	}
	benchEvaluator(f, b)
}

func BenchmarkNaiveTreeEvaluation(b *testing.B) {
	flag.Parse()
	f := func(forest *pb.Forest) Evaluator {
		return &ForestEvaluator{forest}
	}
	benchEvaluator(f, b)
}

func BenchmarkParallelTreeEvaluation(b *testing.B) {
	flag.Parse()
	f := func(forest *pb.Forest) Evaluator {
		return NewParallelForestEvaluator(forest)
	}
	benchEvaluator(f, b)
}