// SPDX-License-Identifier: Unlicense OR MIT

package block

import (
	"gioui.org/layout"
	"gioui.org/op"
	"image"
	"testing"
)

func TestLine_Layout(t *testing.T) {
	tests := []struct {
		name         string
		constraints  layout.Constraints
		line         Line
		segments     []Segment
		wantSize     image.Point
		wantBaseline int
	}{
		{
			name: "AxisHorizontalOneSegment",
			constraints: layout.Constraints{
				Max: image.Point{X: 128, Y: 128},
			},
			line: Line{
				Axis:     AxisHorizontal,
				Overflow: OverflowClip,
			},
			segments: []Segment{
				{Widget: widget008By008},
			},
			wantSize:     image.Point{X: 8, Y: 8},
			wantBaseline: 4,
		},
		{
			name: "AxisHorizontalTwoSegments",
			constraints: layout.Constraints{
				Max: image.Point{X: 128, Y: 128},
			},
			line: Line{
				Axis:     AxisHorizontal,
				Overflow: OverflowClip,
			},
			segments: []Segment{
				{Widget: widget008By008},
				{Widget: widget016By016},
			},
			wantSize:     image.Point{X: 24, Y: 20},
			wantBaseline: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gtx := layout.Context{
				Constraints: tt.constraints,
				Locale:      english,
				Ops:         new(op.Ops),
			}
			got := tt.line.Layout(gtx, tt.segments...)
			if got.Size != tt.wantSize {
				t.Errorf("got %v, want wantSize %v", got.Size, tt.wantSize)
			}
			if got.Baseline != tt.wantBaseline {
				t.Errorf("got %v, want wantBaseline %v", got.Baseline, tt.wantBaseline)
			}
		})
	}
}

func TestLine_validateConstraints(t *testing.T) {
	tests := []struct {
		name        string
		constraints layout.Constraints
		line        Line
		segments    []Segment
		wantErr     bool
	}{
		{
			name: "None",
			constraints: layout.Constraints{
				Max: image.Point{X: 128, Y: 128},
			},
			line: Line{
				Axis:     AxisHorizontal,
				Overflow: OverflowClip,
			},
			segments: []Segment{
				{Widget: widget008By008},
			},
			wantErr: false,
		},
		{
			name: "NoSegments",
			constraints: layout.Constraints{
				Max: image.Point{X: 128, Y: 128},
			},
			line: Line{
				Axis:     AxisHorizontal,
				Overflow: OverflowClip,
			},
			segments: []Segment{},
			wantErr:  true,
		},
		{
			name: "NoHorizontalSpace",
			constraints: layout.Constraints{
				Max: image.Point{X: 0, Y: 128},
			},
			line: Line{
				Axis:     AxisHorizontal,
				Overflow: OverflowClip,
			},
			segments: []Segment{
				{Widget: widget008By008},
			},
			wantErr: true,
		},
		{
			name: "NoVerticalSpace",
			constraints: layout.Constraints{
				Max: image.Point{X: 128, Y: 0},
			},
			line: Line{
				Axis:     AxisHorizontal,
				Overflow: OverflowClip,
			},
			segments: []Segment{
				{Widget: widget008By008},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gtx := layout.Context{
				Constraints: tt.constraints,
				Ops:         new(op.Ops),
			}
			gotErr := tt.line.validateConstraints(gtx, tt.segments)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("gotErr = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestLine_layoutHorizontal_LTR(t *testing.T) {
	viewportSize := image.Point{X: 128, Y: 128}
	tests := []struct {
		name           string
		line           Line
		segments       []Segment
		wantDimensions []layout.Dimensions
		wantOffsets    []image.Point
		wantSize       image.Point
		wantBaseline   int
	}{
		{
			name: "OneSmallSquare",
			line: Line{Axis: AxisHorizontal, Overflow: OverflowClip, Expand: false},
			segments: []Segment{
				{Widget: widget008By008},
			},
			wantDimensions: []layout.Dimensions{
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
			},
			wantOffsets: []image.Point{
				{X: 0, Y: 0},
			},
			wantSize:     image.Point{X: 8, Y: 8},
			wantBaseline: 4,
		},
		{
			name: "OneSmallSquareExpand",
			line: Line{Axis: AxisHorizontal, Overflow: OverflowClip, Expand: true},
			segments: []Segment{
				{Widget: widget008By008},
			},
			wantDimensions: []layout.Dimensions{
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
			},
			wantOffsets: []image.Point{
				{X: 0, Y: 0},
			},
			wantSize:     image.Point{X: 128, Y: 8},
			wantBaseline: 4,
		},
		{
			name: "TwoSmallSquares",
			line: Line{Axis: AxisHorizontal, Overflow: OverflowClip, Expand: false},
			segments: []Segment{
				{Widget: widget008By008},
				{Widget: widget008By008},
			},
			wantDimensions: []layout.Dimensions{
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
			},
			wantOffsets: []image.Point{
				{X: 0, Y: 0},
				{X: 8, Y: 0},
			},
			wantSize:     image.Point{X: 16, Y: 8},
			wantBaseline: 4,
		},
		{
			name: "TwoSmallSquaresExpand",
			line: Line{Axis: AxisHorizontal, Overflow: OverflowClip, Expand: true},
			segments: []Segment{
				{Widget: widget008By008},
				{Widget: widget008By008},
			},
			wantDimensions: []layout.Dimensions{
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
			},
			wantOffsets: []image.Point{
				{X: 0, Y: 0},
				{X: 8, Y: 0},
			},
			wantSize:     image.Point{X: 128, Y: 8},
			wantBaseline: 4,
		},
		{
			name: "TwoDifferentSquares",
			line: Line{Axis: AxisHorizontal, Overflow: OverflowClip, Expand: false},
			segments: []Segment{
				{Widget: widget008By008},
				{Widget: widget016By016},
			},
			wantDimensions: []layout.Dimensions{
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
				{Size: image.Point{X: 16, Y: 16}, Baseline: 0},
			},
			wantOffsets: []image.Point{
				{X: 0, Y: 12},
				{X: 8, Y: 0},
			},
			wantSize:     image.Point{X: 24, Y: 20},
			wantBaseline: 4,
		},
		{
			name: "TwoDifferentSquaresExpand",
			line: Line{Axis: AxisHorizontal, Overflow: OverflowClip, Expand: true},
			segments: []Segment{
				{Widget: widget008By008},
				{Widget: widget016By016},
			},
			wantDimensions: []layout.Dimensions{
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
				{Size: image.Point{X: 16, Y: 16}, Baseline: 0},
			},
			wantOffsets: []image.Point{
				{X: 0, Y: 12},
				{X: 8, Y: 0},
			},
			wantSize:     image.Point{X: 128, Y: 20},
			wantBaseline: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.wantDimensions) != len(tt.segments) {
				t.Errorf("got %v, want dimensions count %v", len(tt.wantDimensions), len(tt.segments))
			}
			if len(tt.wantOffsets) != len(tt.segments) {
				t.Errorf("got %v, want offsets count %v", len(tt.wantOffsets), len(tt.segments))
			}
			gtx := layout.Context{
				Constraints: layout.Constraints{Max: viewportSize},
				Locale:      english,
				Ops:         new(op.Ops),
			}
			got, rSegments := tt.line.layoutHorizontalClip(gtx, tt.segments)
			if len(rSegments) != len(tt.segments) {
				t.Errorf("got %v, want segments count %v", len(rSegments), len(tt.segments))
			}
			if got.Size != tt.wantSize {
				t.Errorf("got %v, want wantSize %v", got.Size, tt.wantSize)
			}
			if got.Baseline != tt.wantBaseline {
				t.Errorf("got %v, want wantBaseline %v", got.Baseline, tt.wantBaseline)
			}
			for index, segment := range rSegments {
				if segment.dimensions.Size != tt.wantDimensions[index].Size {
					t.Errorf("element %d got Size %v, expected %v", index, segment.dimensions.Size, tt.wantDimensions[index].Size)
				}
				if segment.dimensions.Baseline != tt.wantDimensions[index].Baseline {
					t.Errorf("element %d got Baseline %v, expected %v", index, segment.dimensions.Baseline, tt.wantDimensions[index].Baseline)
				}
				if segment.offset != tt.wantOffsets[index] {
					t.Errorf("element %d got offset %v, expected %v", index, segment.offset, tt.wantOffsets[index])
				}
			}
		})
	}
}

func TestLine_layoutHorizontal_RTL(t *testing.T) {
	viewportSize := image.Point{X: 128, Y: 128}
	tests := []struct {
		name           string
		line           Line
		segments       []Segment
		wantDimensions []layout.Dimensions
		wantOffsets    []image.Point
		wantSize       image.Point
		wantBaseline   int
	}{
		{
			name: "OneSmallSquare",
			line: Line{Axis: AxisHorizontal, Overflow: OverflowClip, Expand: false},
			segments: []Segment{
				{Widget: widget008By008},
			},
			wantDimensions: []layout.Dimensions{
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
			},
			wantOffsets: []image.Point{
				{X: 0, Y: 0},
			},
			wantSize:     image.Point{X: 8, Y: 8},
			wantBaseline: 4,
		},
		{
			name: "OneSmallSquareExpand",
			line: Line{Axis: AxisHorizontal, Overflow: OverflowClip, Expand: true},
			segments: []Segment{
				{Widget: widget008By008},
			},
			wantDimensions: []layout.Dimensions{
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
			},
			wantOffsets: []image.Point{
				{X: 120, Y: 0},
			},
			wantSize:     image.Point{X: 128, Y: 8},
			wantBaseline: 4,
		},
		{
			name: "TwoSmallSquares",
			line: Line{Axis: AxisHorizontal, Overflow: OverflowClip, Expand: false},
			segments: []Segment{
				{Widget: widget008By008},
				{Widget: widget008By008},
			},
			wantDimensions: []layout.Dimensions{
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
			},
			wantOffsets: []image.Point{
				{X: 8, Y: 0},
				{X: 0, Y: 0},
			},
			wantSize:     image.Point{X: 16, Y: 8},
			wantBaseline: 4,
		},
		{
			name: "TwoSmallSquaresExpand",
			line: Line{Axis: AxisHorizontal, Overflow: OverflowClip, Expand: true},
			segments: []Segment{
				{Widget: widget008By008},
				{Widget: widget008By008},
			},
			wantDimensions: []layout.Dimensions{
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
			},
			wantOffsets: []image.Point{
				{X: 120, Y: 0},
				{X: 112, Y: 0},
			},
			wantSize:     image.Point{X: 128, Y: 8},
			wantBaseline: 4,
		},
		{
			name: "TwoDifferentSquares",
			line: Line{Axis: AxisHorizontal, Overflow: OverflowClip, Expand: false},
			segments: []Segment{
				{Widget: widget008By008},
				{Widget: widget016By016},
			},
			wantDimensions: []layout.Dimensions{
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
				{Size: image.Point{X: 16, Y: 16}, Baseline: 0},
			},
			wantOffsets: []image.Point{
				{X: 16, Y: 12},
				{X: 0, Y: 0},
			},
			wantSize:     image.Point{X: 24, Y: 20},
			wantBaseline: 4,
		},
		{
			name: "TwoDifferentSquaresExpand",
			line: Line{Axis: AxisHorizontal, Overflow: OverflowClip, Expand: true},
			segments: []Segment{
				{Widget: widget008By008},
				{Widget: widget016By016},
			},
			wantDimensions: []layout.Dimensions{
				{Size: image.Point{X: 8, Y: 8}, Baseline: 4},
				{Size: image.Point{X: 16, Y: 16}, Baseline: 0},
			},
			wantOffsets: []image.Point{
				{X: 120, Y: 12},
				{X: 104, Y: 0},
			},
			wantSize:     image.Point{X: 128, Y: 20},
			wantBaseline: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.wantDimensions) != len(tt.segments) {
				t.Errorf("got %v, want dimensions count %v", len(tt.wantDimensions), len(tt.segments))
			}
			if len(tt.wantOffsets) != len(tt.segments) {
				t.Errorf("got %v, want offsets count %v", len(tt.wantOffsets), len(tt.segments))
			}
			gtx := layout.Context{
				Constraints: layout.Constraints{Max: viewportSize},
				Locale:      yidish,
				Ops:         new(op.Ops),
			}
			got, rSegments := tt.line.layoutHorizontalClip(gtx, tt.segments)
			if len(rSegments) != len(tt.segments) {
				t.Errorf("got %v, want segments count %v", len(rSegments), len(tt.segments))
			}
			if got.Size != tt.wantSize {
				t.Errorf("got %v, want wantSize %v", got.Size, tt.wantSize)
			}
			if got.Baseline != tt.wantBaseline {
				t.Errorf("got %v, want wantBaseline %v", got.Baseline, tt.wantBaseline)
			}
			for index, segment := range rSegments {
				if segment.dimensions.Size != tt.wantDimensions[index].Size {
					t.Errorf("element %d got Size %v, expected %v", index, segment.dimensions.Size, tt.wantDimensions[index].Size)
				}
				if segment.dimensions.Baseline != tt.wantDimensions[index].Baseline {
					t.Errorf("element %d got Baseline %v, expected %v", index, segment.dimensions.Baseline, tt.wantDimensions[index].Baseline)
				}
				if segment.offset != tt.wantOffsets[index] {
					t.Errorf("element %d got offset %v, expected %v", index, segment.offset, tt.wantOffsets[index])
				}
			}
		})
	}
}
