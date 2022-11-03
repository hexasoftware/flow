package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/hexasoftware/flow"
	"github.com/hexasoftware/flow/example/demos/ops/ml"
	"gonum.org/v1/gonum/mat"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write mem profile to file")
)

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer pprof.WriteHeapProfile(f)
	}
	// Registry for machine learning
	r := ml.New()

	f := flow.New()
	f.UseRegistry(r)

	samples := []float64{
		0, 0,
		0, 1,
		1, 0,
		1, 1,
	}
	labels := []float64{
		0,
		1,
		1,
		0,
	}
	learningRate := float64(0.3)

	nInputs := 2
	nHidden := 5
	nOutput := 1
	nSamples := 4

	matSamples := mat.NewDense(nSamples, 2, samples)
	matLabels := mat.NewDense(nSamples, 1, labels)

	// Define input
	// Make a matrix out of the input and output
	x := f.In(0)
	y := f.In(1)

	wHidden := f.Var("wHidden", f.Op("matNewRand", nInputs, nHidden))

	wOut := f.Var("wOut", f.Op("matNewRand", nHidden, nOutput))

	// Forward process
	hiddenLayerInput := f.Op("matMul", x, wHidden)
	hiddenLayerActivations := f.Op("matSigmoid", hiddenLayerInput)
	outputLayerInput := f.Op("matMul", hiddenLayerActivations, wOut)
	// Activations
	output := f.Op("matSigmoid", outputLayerInput)

	// Back propagation
	// output weights
	networkError := f.Op("matSub", y, output)
	slopeOutputLayer := f.Op("matSigmoidPrime", output)
	dOutput := f.Op("matMulElem", networkError, slopeOutputLayer)
	wOutAdj := f.Op("matScale",
		learningRate,
		f.Op("matMul", f.Op("matTranspose", hiddenLayerActivations), dOutput),
	)

	// hidden weights
	errorAtHiddenLayer := f.Op("matMul", dOutput, f.Op("matTranspose", wOut))
	slopeHiddenLayer := f.Op("matSigmoidPrime", hiddenLayerActivations)
	dHiddenLayer := f.Op("matMulElem", errorAtHiddenLayer, slopeHiddenLayer)
	wHiddenAdj := f.Op("matScale",
		learningRate,
		f.Op("matMul", f.Op("matTranspose", x), dHiddenLayer),
	)

	// Adjust the parameters
	setwOut := f.SetVar("wOut", f.Op("matAdd", wOut, wOutAdj))
	setwHidden := f.SetVar("wHidden", f.Op("matAdd", wHidden, wHiddenAdj))

	// Training
	for i := 0; i < 5000; i++ {
		sess := f.NewSession()
		sess.Inputs(matSamples, matLabels)
		_, err := sess.Run(setwOut, setwHidden)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Same as above because its simple
	// Usually it retains different data
	testSamples := matSamples
	testLabels := matLabels

	res, err := output.Process(testSamples)
	if err != nil {
		log.Fatal(err)
	}

	predictions := res.(mat.Matrix)
	log.Println("Predictions", predictions)

	var rights int
	numPreds, _ := predictions.Dims()
	log.Println("Number of predictions:", numPreds)
	for i := 0; i < numPreds; i++ {
		if predictions.At(i, 0) > 0.5 && testLabels.At(i, 0) == 1.0 ||
			predictions.At(i, 0) < 0.5 && testLabels.At(i, 0) == 0 {
			rights++
		}
	}

	accuracy := float64(rights) / float64(numPreds)
	fmt.Printf("\nAccuracy = %0.2f\n\n", accuracy)
}
