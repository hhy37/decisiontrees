package decisiontrees

import (
	"fmt"
	"math"
	"sort"
)

type LabelledPrediction struct {
	Label      bool
	Prediction float64
}

type LabelledPredictions []LabelledPrediction

func (l LabelledPredictions) Len() int {
	return len(l)
}

func (l LabelledPredictions) Swap(i int, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l LabelledPredictions) Less(i int, j int) bool {
	return l[i].Prediction < l[j].Prediction
}

func (l LabelledPredictions) ROC() float64 {
	sort.Sort(l)
	numPositives, numNegatives, weightedSum := 0, 0, 0
	for _, e := range l {
		if e.Label {
			numPositives += 1
		} else {
			numNegatives += 1
			weightedSum += numPositives
		}
	}
	return float64(weightedSum) / float64(numPositives*numNegatives)
}

func (l LabelledPredictions) String() string {
	return fmt.Sprintf(
		"Size: %v\nROC: %v\nCalibration: %v\nNormalized Entropy: %v\nPositives: %v",
		l.Len(),
		l.ROC(),
		l.Calibration(),
		l.NormalizedEntropy(),
		l.numPositives())
}

func (l LabelledPredictions) numPositives() int {
	s := 0
	for _, e := range l {
		if e.Label {
			s += 1
		}
	}
	return s
}

func (l LabelledPredictions) LogScore() float64 {
	cumulativeLogLoss := 0.0
	for _, e := range l {
		if e.Label {
			cumulativeLogLoss += math.Log2(e.Prediction)
		} else {
			cumulativeLogLoss += math.Log2(1 - e.Prediction)
		}
	}
	return cumulativeLogLoss / float64(l.Len())
}

func (l LabelledPredictions) Calibration() float64 {
	numPositives, sumPredictions := 0, 0.0
	for _, e := range l {
		sumPredictions += e.Prediction
		if e.Label {
			numPositives += 1
		}
	}
	return float64(sumPredictions) / float64(numPositives)
}

func (l LabelledPredictions) NormalizedEntropy() float64 {
	numPositives := 0
	for _, e := range l {
		if e.Label {
			numPositives += 1
		}
	}
	p := float64(numPositives) / float64(l.Len())
	return l.LogScore() / (p*math.Log2(p) + (1-p)*math.Log2(1-p))
}