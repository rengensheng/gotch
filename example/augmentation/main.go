package main

import (
	"fmt"

	"github.com/sugarme/gotch"
	"github.com/sugarme/gotch/vision"
	"github.com/sugarme/gotch/vision/aug"
)

func main() {
	n := 360
	for i := 1; i <= n; i++ {
		img, err := vision.Load("./bb.png")
		if err != nil {
			panic(err)
		}

		// device := gotch.CudaIfAvailable()
		device := gotch.CPU
		imgTs := img.MustTo(device, true)
		// t, err := aug.Compose(aug.WithResize(512, 512)) // NOTE. WithResize just works on CPU.
		// t, err := aug.Compose(aug.WithRandRotate(0, 360), aug.WithColorJitter(0.3, 0.3, 0.3, 0.4))
		// t, err := aug.Compose(aug.WithGaussianBlur([]int64{5, 5}, []float64{1.0, 2.0}), aug.WithRandRotate(0, 360), aug.WithColorJitter(0.3, 0.3, 0.3, 0.3))
		// t, err := aug.Compose(aug.WithRandomCrop([]int64{320, 320}, []int64{10, 10}, true, "constant"))
		// t, err := aug.Compose(aug.WithCenterCrop([]int64{320, 320}))
		// t, err := aug.Compose(aug.WithRandomCutout(aug.WithCutoutValue([]int64{124, 96, 255}), aug.WithCutoutScale([]float64{0.01, 0.1}), aug.WithCutoutRatio([]float64{0.5, 0.5})))
		// t, err := aug.Compose(aug.WithRandomPerspective(aug.WithPerspectiveScale(0.6), aug.WithPerspectivePvalue(0.8)))
		// t, err := aug.Compose(aug.WithRandomAffine(aug.WithAffineDegree([]int64{0, 15}), aug.WithAffineShear([]float64{0, 15})))
		// t, err := aug.Compose(aug.WithRandomGrayscale(0.5))
		// t, err := aug.Compose(aug.WithRandomSolarize(aug.WithSolarizeThreshold(125), aug.WithSolarizePvalue(0.5)))
		// t, err := aug.Compose(aug.WithRandomInvert(0.5))
		// t, err := aug.Compose(aug.WithRandomPosterize(aug.WithPosterizeBits(2), aug.WithPosterizePvalue(1.0)))
		// t, err := aug.Compose(aug.WithRandomAutocontrast())
		// t, err := aug.Compose(aug.WithRandomAdjustSharpness(aug.WithSharpnessPvalue(0.3), aug.WithSharpnessFactor(10)))
		// t, err := aug.Compose(aug.WithRandomEqualize(1.0))
		// t, err := aug.Compose(aug.WithNormalize(aug.WithNormalizeMean([]float64{0.485, 0.456, 0.406}), aug.WithNormalizeStd([]float64{0.229, 0.224, 0.225})))

		t, err := aug.Compose(
			aug.WithResize(200, 200),
			aug.WithRandomVFlip(0.5),
			aug.WithRandomHFlip(0.5),
			aug.WithRandomCutout(),
			aug.OneOf(
				0.3,
				aug.WithColorJitter(0.3, 0.3, 0.3, 0.4),
				aug.WithRandomGrayscale(1.0),
			),
			aug.OneOf(
				0.3,
				aug.WithGaussianBlur([]int64{5, 5}, []float64{1.0, 2.0}),
				aug.WithRandomAffine(),
			),
		)
		if err != nil {
			panic(err)
		}

		out := t.Transform(imgTs)
		fname := fmt.Sprintf("./output/bb-%03d.png", i)
		err = vision.Save(out, fname)
		if err != nil {
			panic(err)
		}
		imgTs.MustDrop()
		out.MustDrop()

		fmt.Printf("%03d/%v completed.\n", i, n)
	}
}
