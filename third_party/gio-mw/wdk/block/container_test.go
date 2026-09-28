// SPDX-License-Identifier: Unlicense OR MIT

package block

import (
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"image"
	"testing"
)

func TestContainer_validateConstraints(t *testing.T) {
	tests := []struct {
		name        string
		constraints layout.Constraints
		container   Container
		wantErr     bool
	}{
		{
			name:        "Valid",
			constraints: constraintsDefault,
			container:   Container{},
			wantErr:     false,
		},
		{
			name:        "ZeroHeightConstraint",
			constraints: constraintsZeroHeight,
			container:   Container{},
			wantErr:     true,
		},
		{
			name:        "ZeroWidthConstraint",
			constraints: constraintsZeroWidth,
			container:   Container{},
			wantErr:     true,
		},
		{
			name:        "NegativeMinSizeX",
			constraints: constraintsDefault,
			container: Container{
				MinSize: image.Point{X: -1, Y: 0},
			},
			wantErr: true,
		},
		{
			name:        "NegativeMinSizeY",
			constraints: constraintsDefault,
			container: Container{
				MinSize: image.Point{X: 0, Y: -1},
			},
			wantErr: true,
		},
		{
			name:        "NegativeMaxSizeX",
			constraints: constraintsDefault,
			container: Container{
				MaxSize: image.Point{X: -1, Y: 0},
			},
			wantErr: true,
		},
		{
			name:        "NegativeMaxSizeY",
			constraints: constraintsDefault,
			container: Container{
				MaxSize: image.Point{X: 0, Y: -1},
			},
			wantErr: true,
		},
		{
			name:        "MinSizeXGreaterThanGtxMaxX",
			constraints: constraintsSmall,
			container: Container{
				MinSize: image.Point{X: 64, Y: 0},
			},
			wantErr: true,
		},
		{
			name:        "MinSizeYGreaterThanGtxMaxY",
			constraints: constraintsSmall,
			container: Container{
				MinSize: image.Point{X: 0, Y: 64},
			},
			wantErr: true,
		},
		{
			name:        "MinSizeXGreaterThanMaxSizeX",
			constraints: constraintsDefault,
			container: Container{
				MinSize: image.Point{X: 64, Y: 0},
				MaxSize: image.Point{X: 32, Y: 0},
			},
			wantErr: true,
		},
		{
			name:        "MinSizeYGreaterThanMaxSizeY",
			constraints: constraintsDefault,
			container: Container{
				MinSize: image.Point{X: 0, Y: 64},
				MaxSize: image.Point{X: 0, Y: 32},
			},
			wantErr: true,
		},
		{
			name:        "FillNoneWithMinSizeX",
			constraints: constraintsDefault,
			container: Container{
				FillDirection: FillNone,
				MinSize:       image.Point{X: 64, Y: 0},
			},
			wantErr: false,
		},
		{
			name:        "FillNoneWithMaxSizeX",
			constraints: constraintsDefault,
			container: Container{
				FillDirection: FillNone,
				MaxSize:       image.Point{X: 64, Y: 0},
			},
			wantErr: false,
		},
		{
			name:        "FillHorizontalWithMinSizeX",
			constraints: constraintsDefault,
			container: Container{
				FillDirection: FillHorizontal,
				MinSize:       image.Point{X: 64, Y: 0},
			},
			wantErr: true,
		},
		{
			name:        "FillHorizontalWithMaxSizeX",
			constraints: constraintsDefault,
			container: Container{
				FillDirection: FillHorizontal,
				MaxSize:       image.Point{X: 64, Y: 0},
			},
			wantErr: true,
		},
		{
			name:        "FillHorizontalWithMaxSizeY",
			constraints: constraintsDefault,
			container: Container{
				FillDirection: FillHorizontal,
				MaxSize:       image.Point{X: 0, Y: 64},
			},
			wantErr: false,
		},
		{
			name:        "FillVerticalWithMinSizeY",
			constraints: constraintsDefault,
			container: Container{
				FillDirection: FillVertical,
				MinSize:       image.Point{X: 0, Y: 64},
			},
			wantErr: true,
		},
		{
			name:        "FillVerticalWithMaxSizeY",
			constraints: constraintsDefault,
			container: Container{
				FillDirection: FillVertical,
				MaxSize:       image.Point{X: 0, Y: 64},
			},
			wantErr: true,
		},
		{
			name:        "FillVerticalWithMaxSizeX",
			constraints: constraintsDefault,
			container: Container{
				FillDirection: FillVertical,
				MaxSize:       image.Point{X: 64, Y: 0},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gtx := layout.Context{
				Constraints: tt.constraints,
			}
			err := tt.container.validateConstraints(gtx)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestContainer_updateConstraints(t *testing.T) {
	tests := []struct {
		name        string
		constraints layout.Constraints
		container   Container
		wantMin     image.Point
		wantMax     image.Point
	}{
		{
			name:        "NoChanges",
			constraints: constraintsDefault,
			container:   Container{},
			wantMin:     image.Point{X: 0, Y: 0},
			wantMax:     image.Point{X: 128, Y: 128},
		},
		{
			name:        "MinSizeX",
			constraints: constraintsDefault,
			container: Container{
				MinSize: image.Point{X: 64, Y: 0},
			},
			wantMin: image.Point{X: 64, Y: 0},
			wantMax: image.Point{X: 128, Y: 128},
		},
		{
			name: "MinSizeXWithMinConstraints",
			constraints: layout.Constraints{
				Min: image.Point{X: 96, Y: 0},
				Max: image.Point{X: 128, Y: 128},
			},
			container: Container{
				MinSize: image.Point{X: 64, Y: 0},
			},
			wantMin: image.Point{X: 96, Y: 0},
			wantMax: image.Point{X: 128, Y: 128},
		},
		{
			name:        "MinSizeY",
			constraints: constraintsDefault,
			container: Container{
				MinSize: image.Point{X: 0, Y: 64},
			},
			wantMin: image.Point{X: 0, Y: 64},
			wantMax: image.Point{X: 128, Y: 128},
		},
		{
			name: "MinSizeYWithMinConstraints",
			constraints: layout.Constraints{
				Min: image.Point{X: 0, Y: 96},
				Max: image.Point{X: 128, Y: 128},
			},
			container: Container{
				MinSize: image.Point{X: 0, Y: 64},
			},
			wantMin: image.Point{X: 0, Y: 96},
			wantMax: image.Point{X: 128, Y: 128},
		},
		{
			name:        "MaxSizeX",
			constraints: constraintsDefault,
			container: Container{
				MaxSize: image.Point{X: 64, Y: 0},
			},
			wantMin: image.Point{X: 0, Y: 0},
			wantMax: image.Point{X: 64, Y: 128},
		},
		{
			name:        "MaxSizeY",
			constraints: constraintsDefault,
			container: Container{
				MaxSize: image.Point{X: 0, Y: 64},
			},
			wantMin: image.Point{X: 0, Y: 0},
			wantMax: image.Point{X: 128, Y: 64},
		},
		{
			name:        "MinSizeXEqualMaxSizeX",
			constraints: constraintsDefault,
			container: Container{
				MinSize: image.Point{X: 64, Y: 0},
				MaxSize: image.Point{X: 64, Y: 0},
			},
			wantMin: image.Point{X: 64, Y: 0},
			wantMax: image.Point{X: 64, Y: 128},
		},
		{
			name:        "MinSizeYEqualMaxSizeY",
			constraints: constraintsDefault,
			container: Container{
				MinSize: image.Point{X: 0, Y: 64},
				MaxSize: image.Point{X: 0, Y: 64},
			},
			wantMin: image.Point{X: 0, Y: 64},
			wantMax: image.Point{X: 128, Y: 64},
		},
		{
			name:        "FillHorizontal",
			constraints: constraintsDefault,
			container: Container{
				FillDirection: FillHorizontal,
			},
			wantMin: image.Point{X: 128, Y: 0},
			wantMax: image.Point{X: 128, Y: 128},
		},
		{
			name:        "FillVertical",
			constraints: constraintsDefault,
			container: Container{
				FillDirection: FillVertical,
			},
			wantMin: image.Point{X: 0, Y: 128},
			wantMax: image.Point{X: 128, Y: 128},
		},
		{
			name:        "FillBoth",
			constraints: constraintsDefault,
			container: Container{
				FillDirection: FillBoth,
			},
			wantMin: image.Point{X: 128, Y: 128},
			wantMax: image.Point{X: 128, Y: 128},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gtx := layout.Context{
				Constraints: tt.constraints,
			}
			got := tt.container.updateConstraints(gtx)
			if got.Constraints.Min != tt.wantMin {
				t.Errorf("Min = %v, want %v", got.Constraints.Min, tt.wantMin)
			}
			if got.Constraints.Max != tt.wantMax {
				t.Errorf("Max = %v, want %v", got.Constraints.Max, tt.wantMax)
			}
		})
	}
}

func TestContainer_getChildXOffset_withFill(t *testing.T) {
	tests := []struct {
		name        string
		locale      system.Locale
		constraints layout.Constraints
		gravities   []Gravity
		childWidth  int
		wantOffset  int
	}{
		{
			name:        "StartLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityMiddleStart, GravityBottomStart},
			childWidth:  16,
			wantOffset:  0,
		},
		{
			name:        "StartRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityMiddleStart, GravityBottomStart},
			childWidth:  16,
			wantOffset:  112,
		},
		{
			name:        "CenterLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopCenter, GravityMiddleCenter, GravityBottomCenter},
			childWidth:  16,
			wantOffset:  56,
		},
		{
			name:        "CenterRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopCenter, GravityMiddleCenter, GravityBottomCenter},
			childWidth:  16,
			wantOffset:  56,
		},
		{
			name:        "EndLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopEnd, GravityMiddleEnd, GravityBottomEnd},
			childWidth:  16,
			wantOffset:  112,
		},
		{
			name:        "EndRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopEnd, GravityMiddleEnd, GravityBottomEnd},
			childWidth:  16,
			wantOffset:  0,
		},
		{
			name:        "BigStartLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityMiddleStart, GravityBottomStart},
			childWidth:  128,
			wantOffset:  0,
		},
		{
			name:        "BigStartRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityMiddleStart, GravityBottomStart},
			childWidth:  128,
			wantOffset:  0,
		},
		{
			name:        "BigCenterLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopCenter, GravityMiddleCenter, GravityBottomCenter},
			childWidth:  128,
			wantOffset:  0,
		},
		{
			name:        "BigCenterRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopCenter, GravityMiddleCenter, GravityBottomCenter},
			childWidth:  128,
			wantOffset:  0,
		},
		{
			name:        "BigEndLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopEnd, GravityMiddleEnd, GravityBottomEnd},
			childWidth:  128,
			wantOffset:  0,
		},
		{
			name:        "BigEndRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopEnd, GravityMiddleEnd, GravityBottomEnd},
			childWidth:  128,
			wantOffset:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gtx := layout.Context{
				Constraints: tt.constraints,
				Locale:      tt.locale,
			}
			for _, gravity := range tt.gravities {
				t.Run(gravity.String(), func(t *testing.T) {
					container := Container{
						Gravity:       gravity,
						FillDirection: FillBoth,
					}
					got := container.getChildXOffset(gtx, tt.childWidth)
					if got != tt.wantOffset {
						t.Errorf("offset = %v, want %v", got, tt.wantOffset)
					}
				})
			}
		})
	}
}

func TestContainer_getChildXOffset_withoutFill(t *testing.T) {
	tests := []struct {
		name        string
		locale      system.Locale
		constraints layout.Constraints
		gravities   []Gravity
		childWidth  int
		wantOffset  int
	}{
		{
			name:        "StartLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityMiddleStart, GravityBottomStart},
			childWidth:  16,
			wantOffset:  0,
		},
		{
			name:        "StartRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityMiddleStart, GravityBottomStart},
			childWidth:  16,
			wantOffset:  0,
		},
		{
			name:        "CenterLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopCenter, GravityMiddleCenter, GravityBottomCenter},
			childWidth:  16,
			wantOffset:  0,
		},
		{
			name:        "CenterRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopCenter, GravityMiddleCenter, GravityBottomCenter},
			childWidth:  16,
			wantOffset:  0,
		},
		{
			name:        "EndLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopEnd, GravityMiddleEnd, GravityBottomEnd},
			childWidth:  16,
			wantOffset:  0,
		},
		{
			name:        "EndRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopEnd, GravityMiddleEnd, GravityBottomEnd},
			childWidth:  16,
			wantOffset:  0,
		},
		{
			name:        "BigStartLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityMiddleStart, GravityBottomStart},
			childWidth:  128,
			wantOffset:  0,
		},
		{
			name:        "BigStartRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityMiddleStart, GravityBottomStart},
			childWidth:  128,
			wantOffset:  0,
		},
		{
			name:        "BigCenterLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopCenter, GravityMiddleCenter, GravityBottomCenter},
			childWidth:  128,
			wantOffset:  0,
		},
		{
			name:        "BigCenterRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopCenter, GravityMiddleCenter, GravityBottomCenter},
			childWidth:  128,
			wantOffset:  0,
		},
		{
			name:        "BigEndLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopEnd, GravityMiddleEnd, GravityBottomEnd},
			childWidth:  128,
			wantOffset:  0,
		},
		{
			name:        "BigEndRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopEnd, GravityMiddleEnd, GravityBottomEnd},
			childWidth:  128,
			wantOffset:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gtx := layout.Context{
				Constraints: tt.constraints,
				Locale:      tt.locale,
			}
			for _, gravity := range tt.gravities {
				t.Run(gravity.String(), func(t *testing.T) {
					container := Container{
						Gravity: gravity,
					}
					got := container.getChildXOffset(gtx, tt.childWidth)
					if got != tt.wantOffset {
						t.Errorf("offset = %v, want %v", got, tt.wantOffset)
					}
				})
			}
		})
	}
}

func TestContainer_getChildYOffset_withFill(t *testing.T) {
	tests := []struct {
		name        string
		locale      system.Locale
		constraints layout.Constraints
		gravities   []Gravity
		childHeight int
		wantOffset  int
	}{
		{
			name:        "TopLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityTopCenter, GravityTopEnd},
			childHeight: 16,
			wantOffset:  0,
		},
		{
			name:        "TopRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityTopCenter, GravityTopEnd},
			childHeight: 16,
			wantOffset:  0,
		},
		{
			name:        "MiddleLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityMiddleStart, GravityMiddleCenter, GravityMiddleEnd},
			childHeight: 16,
			wantOffset:  56,
		},
		{
			name:        "MiddleRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityMiddleStart, GravityMiddleCenter, GravityMiddleEnd},
			childHeight: 16,
			wantOffset:  56,
		},
		{
			name:        "BottomLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityBottomStart, GravityBottomCenter, GravityBottomEnd},
			childHeight: 16,
			wantOffset:  112,
		},
		{
			name:        "BottomRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityBottomStart, GravityBottomCenter, GravityBottomEnd},
			childHeight: 16,
			wantOffset:  112,
		},

		{
			name:        "BigTopLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityTopCenter, GravityTopEnd},
			childHeight: 128,
			wantOffset:  0,
		},
		{
			name:        "BigTopRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityTopCenter, GravityTopEnd},
			childHeight: 128,
			wantOffset:  0,
		},
		{
			name:        "BigMiddleLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityMiddleStart, GravityMiddleCenter, GravityMiddleEnd},
			childHeight: 128,
			wantOffset:  0,
		},
		{
			name:        "BigMiddleRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityMiddleStart, GravityMiddleCenter, GravityMiddleEnd},
			childHeight: 128,
			wantOffset:  0,
		},
		{
			name:        "BigBottomLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityBottomStart, GravityBottomCenter, GravityBottomEnd},
			childHeight: 128,
			wantOffset:  0,
		},
		{
			name:        "BigBottomLTR",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityBottomStart, GravityBottomCenter, GravityBottomEnd},
			childHeight: 128,
			wantOffset:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gtx := layout.Context{
				Constraints: tt.constraints,
				Locale:      tt.locale,
			}
			for _, gravity := range tt.gravities {
				t.Run(gravity.String(), func(t *testing.T) {
					container := Container{
						Gravity:       gravity,
						FillDirection: FillBoth,
					}
					got := container.getChildYOffset(gtx, tt.childHeight)
					if got != tt.wantOffset {
						t.Errorf("offset = %v, want %v", got, tt.wantOffset)
					}
				})
			}
		})
	}
}

func TestContainer_getChildYOffset_withoutFill(t *testing.T) {
	tests := []struct {
		name        string
		locale      system.Locale
		constraints layout.Constraints
		gravities   []Gravity
		childHeight int
		wantOffset  int
	}{
		{
			name:        "TopLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityTopCenter, GravityTopEnd},
			childHeight: 16,
			wantOffset:  0,
		},
		{
			name:        "TopRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityTopCenter, GravityTopEnd},
			childHeight: 16,
			wantOffset:  0,
		},
		{
			name:        "MiddleLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityMiddleStart, GravityMiddleCenter, GravityMiddleEnd},
			childHeight: 16,
			wantOffset:  0,
		},
		{
			name:        "MiddleRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityMiddleStart, GravityMiddleCenter, GravityMiddleEnd},
			childHeight: 16,
			wantOffset:  0,
		},
		{
			name:        "BottomLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityBottomStart, GravityBottomCenter, GravityBottomEnd},
			childHeight: 16,
			wantOffset:  0,
		},
		{
			name:        "BottomRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityBottomStart, GravityBottomCenter, GravityBottomEnd},
			childHeight: 16,
			wantOffset:  0,
		},

		{
			name:        "BigTopLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityTopCenter, GravityTopEnd},
			childHeight: 128,
			wantOffset:  0,
		},
		{
			name:        "BigTopRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityTopStart, GravityTopCenter, GravityTopEnd},
			childHeight: 128,
			wantOffset:  0,
		},
		{
			name:        "BigMiddleLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityMiddleStart, GravityMiddleCenter, GravityMiddleEnd},
			childHeight: 128,
			wantOffset:  0,
		},
		{
			name:        "BigMiddleRTL",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityMiddleStart, GravityMiddleCenter, GravityMiddleEnd},
			childHeight: 128,
			wantOffset:  0,
		},
		{
			name:        "BigBottomLTR",
			locale:      english,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityBottomStart, GravityBottomCenter, GravityBottomEnd},
			childHeight: 128,
			wantOffset:  0,
		},
		{
			name:        "BigBottomLTR",
			locale:      yidish,
			constraints: constraintsDefault,
			gravities:   []Gravity{GravityBottomStart, GravityBottomCenter, GravityBottomEnd},
			childHeight: 128,
			wantOffset:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gtx := layout.Context{
				Constraints: tt.constraints,
				Locale:      tt.locale,
			}
			for _, gravity := range tt.gravities {
				t.Run(gravity.String(), func(t *testing.T) {
					container := Container{
						Gravity: gravity,
					}
					got := container.getChildYOffset(gtx, tt.childHeight)
					if got != tt.wantOffset {
						t.Errorf("offset = %v, want %v", got, tt.wantOffset)
					}
				})
			}
		})
	}
}

func TestContainer_getChildBaseline(t *testing.T) {
	tests := []struct {
		name            string
		constraints     layout.Constraints
		container       Container
		childDimensions layout.Dimensions
		wantBaseline    int
	}{
		{
			name:        "GravityTopStart",
			constraints: constraintsDefault,
			container: Container{
				Gravity: GravityTopStart,
			},
			childDimensions: layout.Dimensions{Size: image.Point{X: 16, Y: 16}, Baseline: 8},
			wantBaseline:    120,
		},
		{
			name:        "GravityTopCenter",
			constraints: constraintsDefault,
			container: Container{
				Gravity: GravityTopCenter,
			},
			childDimensions: layout.Dimensions{Size: image.Point{X: 16, Y: 16}, Baseline: 8},
			wantBaseline:    120,
		},
		{
			name:        "GravityTopEnd",
			constraints: constraintsDefault,
			container: Container{
				Gravity: GravityTopEnd,
			},
			childDimensions: layout.Dimensions{Size: image.Point{X: 16, Y: 16}, Baseline: 8},
			wantBaseline:    120,
		},
		{
			name:        "GravityMiddleStart",
			constraints: constraintsDefault,
			container: Container{
				Gravity: GravityMiddleStart,
			},
			childDimensions: layout.Dimensions{Size: image.Point{X: 16, Y: 16}, Baseline: 8},
			wantBaseline:    64,
		},
		{
			name:        "GravityMiddleCenter",
			constraints: constraintsDefault,
			container: Container{
				Gravity: GravityMiddleCenter,
			},
			childDimensions: layout.Dimensions{Size: image.Point{X: 16, Y: 16}, Baseline: 8},
			wantBaseline:    64,
		},
		{
			name:        "GravityMiddleEnd",
			constraints: constraintsDefault,
			container: Container{
				Gravity: GravityMiddleEnd,
			},
			childDimensions: layout.Dimensions{Size: image.Point{X: 16, Y: 16}, Baseline: 8},
			wantBaseline:    64,
		},
		{
			name:        "GravityBottomStart",
			constraints: constraintsDefault,
			container: Container{
				Gravity: GravityBottomStart,
			},
			childDimensions: layout.Dimensions{Size: image.Point{X: 16, Y: 16}, Baseline: 8},
			wantBaseline:    8,
		},
		{
			name:        "GravityBottomCenter",
			constraints: constraintsDefault,
			container: Container{
				Gravity: GravityBottomCenter,
			},
			childDimensions: layout.Dimensions{Size: image.Point{X: 16, Y: 16}, Baseline: 8},
			wantBaseline:    8,
		},
		{
			name:        "GravityBottomEnd",
			constraints: constraintsDefault,
			container: Container{
				Gravity: GravityBottomEnd,
			},
			childDimensions: layout.Dimensions{Size: image.Point{X: 16, Y: 16}, Baseline: 8},
			wantBaseline:    8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			availableHeight := tt.constraints.Max.Y
			got := tt.container.getChildBaseline(tt.childDimensions, availableHeight)
			if got != tt.wantBaseline {
				t.Errorf("baseline = %v, want %v", got, tt.wantBaseline)
			}
		})
	}
}

func TestContainer_clipOverflow(t *testing.T) {
	tests := []struct {
		name         string
		constraints  layout.Constraints
		container    Container
		dimensions   layout.Dimensions
		wantSize     image.Point
		wantBaseline int
	}{
		{
			name:         "NoClipping",
			constraints:  constraintsDefault,
			container:    Container{},
			dimensions:   layout.Dimensions{Size: image.Point{X: 16, Y: 16}, Baseline: 8},
			wantSize:     image.Point{X: 16, Y: 16},
			wantBaseline: 8,
		},
		{
			name:        "ClipWidth",
			constraints: constraintsDefault,
			container: Container{
				MaxSize: image.Point{X: 32, Y: 0},
			},
			dimensions:   layout.Dimensions{Size: image.Point{X: 64, Y: 16}, Baseline: 8},
			wantSize:     image.Point{X: 32, Y: 16},
			wantBaseline: 8,
		},
		{
			name:        "ClipHeight",
			constraints: constraintsDefault,
			container: Container{
				MaxSize: image.Point{X: 0, Y: 32},
			},
			dimensions:   layout.Dimensions{Size: image.Point{X: 16, Y: 64}, Baseline: 48},
			wantSize:     image.Point{X: 16, Y: 32},
			wantBaseline: 16,
		},
		{
			name:        "ClipBoth",
			constraints: constraintsDefault,
			container: Container{
				MaxSize: image.Point{X: 32, Y: 32},
			},
			dimensions:   layout.Dimensions{Size: image.Point{X: 64, Y: 64}, Baseline: 48},
			wantSize:     image.Point{X: 32, Y: 32},
			wantBaseline: 16,
		},
		{
			name:        "ClipUpdateBaseline",
			constraints: constraintsDefault,
			container: Container{
				MaxSize: image.Point{X: 0, Y: 32},
			},
			dimensions:   layout.Dimensions{Size: image.Point{X: 16, Y: 64}, Baseline: 8},
			wantSize:     image.Point{X: 16, Y: 32},
			wantBaseline: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gtx := layout.Context{
				Ops:         new(op.Ops),
				Constraints: tt.constraints,
			}
			callOp := op.Record(gtx.Ops).Stop()
			gotCallOp, gotDims := tt.container.clipOverflow(gtx, callOp, tt.dimensions)
			if gotDims.Size != tt.wantSize {
				t.Errorf("size = %v, want %v", gotDims.Size, tt.wantSize)
			}
			if gotDims.Baseline != tt.wantBaseline {
				t.Errorf("baseline = %v, want %v", gotDims.Baseline, tt.wantBaseline)
			}
			if gotCallOp == (op.CallOp{}) {
				t.Error("expected non-empty CallOp")
			}
		})
	}
}

func TestContainer_layoutWidget(t *testing.T) {
	tests := []struct {
		name         string
		constraints  layout.Constraints
		container    Container
		widget       layout.Widget
		wantSize     image.Point
		wantOffset   image.Point
		wantBaseline int
	}{
		{
			name:        "DefaultLayout",
			constraints: constraintsDefault,
			container: Container{
				FillDirection: FillBoth,
			},
			widget: func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{
					Size:     image.Point{X: 16, Y: 16},
					Baseline: 8,
				}
			},
			wantSize:     image.Point{X: 128, Y: 128},
			wantOffset:   image.Point{X: 0, Y: 0},
			wantBaseline: 120,
		},
		{
			name:        "MiddleCenterLayout",
			constraints: constraintsDefault,
			container: Container{
				Gravity:       GravityMiddleCenter,
				FillDirection: FillBoth,
			},
			widget: func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{
					Size:     image.Point{X: 16, Y: 16},
					Baseline: 8,
				}
			},
			wantSize:     image.Point{X: 128, Y: 128},
			wantOffset:   image.Point{X: 56, Y: 56},
			wantBaseline: 64,
		},
		{
			name:        "BottomEndLayout",
			constraints: constraintsDefault,
			container: Container{
				Gravity:       GravityBottomEnd,
				FillDirection: FillBoth,
			},
			widget: func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{
					Size:     image.Point{X: 16, Y: 16},
					Baseline: 8,
				}
			},
			wantSize:     image.Point{X: 128, Y: 128},
			wantOffset:   image.Point{X: 112, Y: 112},
			wantBaseline: 8,
		},
		{
			name: "WithMinSize",
			constraints: layout.Constraints{
				Min: image.Point{X: 64, Y: 64},
				Max: image.Point{X: 256, Y: 256},
			},
			container: Container{
				Gravity:       GravityBottomEnd,
				FillDirection: FillHorizontal,
			},
			widget: func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{
					Size:     image.Point{X: 16, Y: gtx.Constraints.Min.Y},
					Baseline: 8,
				}
			},
			wantSize:     image.Point{X: 256, Y: 64},
			wantOffset:   image.Point{X: 240, Y: 64},
			wantBaseline: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gtx := layout.Context{
				Ops:         new(op.Ops),
				Constraints: tt.constraints,
			}
			dims, offset := tt.container.layoutWidget(gtx, tt.widget)
			if dims.Size != tt.wantSize {
				t.Errorf("size = %v, want %v", dims.Size, tt.wantSize)
			}
			if offset != tt.wantOffset {
				t.Errorf("offset = %v, want %v", offset, tt.wantOffset)
			}
			if dims.Baseline != tt.wantBaseline {
				t.Errorf("baseline = %v, want %v", dims.Baseline, tt.wantBaseline)
			}
		})
	}
}

func TestContainer_Layout(t *testing.T) {
	smallWidget := func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Point{X: 16, Y: 16}, Baseline: 8}
	}

	tests := []struct {
		name           string
		locale         system.Locale
		constraints    layout.Constraints
		container      Container
		wantDimensions layout.Dimensions
	}{
		{
			name:        "DefaultLayout",
			locale:      english,
			constraints: constraintsDefault,
			container:   Container{},
			wantDimensions: layout.Dimensions{
				Size:     image.Point{X: 16, Y: 16},
				Baseline: 8,
			},
		},
		{
			name:        "FillHorizontalLayout",
			locale:      english,
			constraints: constraintsDefault,
			container: Container{
				FillDirection: FillHorizontal,
			},
			wantDimensions: layout.Dimensions{
				Size:     image.Point{X: 128, Y: 16},
				Baseline: 8,
			},
		},
		{
			name:        "FillVerticalLayout",
			locale:      english,
			constraints: constraintsDefault,
			container: Container{
				FillDirection: FillVertical,
			},
			wantDimensions: layout.Dimensions{
				Size:     image.Point{X: 16, Y: 128},
				Baseline: 120,
			},
		},
		{
			name:        "FillBothLayout",
			locale:      english,
			constraints: constraintsDefault,
			container: Container{
				FillDirection: FillBoth,
			},
			wantDimensions: layout.Dimensions{
				Size:     image.Point{X: 128, Y: 128},
				Baseline: 120,
			},
		},
		{
			name:        "MinSizeLayout",
			locale:      english,
			constraints: constraintsDefault,
			container: Container{
				MinSize: image.Point{X: 32, Y: 32},
			},
			wantDimensions: layout.Dimensions{
				Size:     image.Point{X: 32, Y: 32},
				Baseline: 24,
			},
		},
		{
			name:        "MaxSizeLayout",
			locale:      english,
			constraints: constraintsDefault,
			container: Container{
				MaxSize: image.Point{X: 32, Y: 32},
			},
			wantDimensions: layout.Dimensions{
				Size:     image.Point{X: 16, Y: 16},
				Baseline: 8,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gtx := layout.Context{
				Ops:         new(op.Ops),
				Constraints: tt.constraints,
				Locale:      tt.locale,
			}
			got := tt.container.Layout(gtx, smallWidget)
			if got.Size != tt.wantDimensions.Size {
				t.Errorf("size = %v, want %v", got.Size, tt.wantDimensions.Size)
			}
			if got.Baseline != tt.wantDimensions.Baseline {
				t.Errorf("baseline = %v, want %v", got.Baseline, tt.wantDimensions.Baseline)
			}
		})
	}
}
