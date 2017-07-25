// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vision

import (
	"image/color"
	"math"

	"cloud.google.com/go/internal/version"
	vkit "cloud.google.com/go/vision/apiv1"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
	cpb "google.golang.org/genproto/googleapis/type/color"
)

// Scope is the OAuth2 scope required by the Google Cloud Vision API.
const Scope = "https://www.googleapis.com/auth/cloud-platform"

// Client is a Google Cloud Vision API client.
type Client struct {
	client *vkit.ImageAnnotatorClient
}

// NewClient creates a new vision client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	c, err := vkit.NewImageAnnotatorClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	c.SetGoogleClientInfo("gccl", version.Repo)
	return &Client{client: c}, nil
}

// Close closes the client.
func (c *Client) Close() error {
	return c.client.Close()
}

// Annotate annotates multiple images, each with a potentially differeent set
// of features.
func (c *Client) Annotate(ctx context.Context, requests ...*AnnotateRequest) ([]*Annotations, error) {
	var reqs []*pb.AnnotateImageRequest
	for _, r := range requests {
		reqs = append(reqs, r.toProto())
	}
	res, err := c.client.BatchAnnotateImages(ctx, &pb.BatchAnnotateImagesRequest{Requests: reqs})
	if err != nil {
		return nil, err
	}
	var results []*Annotations
	for _, res := range res.Responses {
		results = append(results, annotationsFromProto(res))
	}
	return results, nil
}

// An AnnotateRequest specifies an image to annotate and the features to look for in that image.
type AnnotateRequest struct {
	// Image is the image to annotate.
	Image *Image
	// MaxFaces is the maximum number of faces to detect in the image.
	// Specifying a number greater than zero enables face detection.
	MaxFaces int
	// MaxLandmarks is the maximum number of landmarks to detect in the image.
	// Specifying a number greater than zero enables landmark detection.
	MaxLandmarks int
	// MaxLogos is the maximum number of logos to detect in the image.
	// Specifying a number greater than zero enables logo detection.
	MaxLogos int
	// MaxLabels is the maximum number of labels to detect in the image.
	// Specifying a number greater than zero enables labels detection.
	MaxLabels int
	// MaxTexts is the maximum number of separate pieces of text to detect in the
	// image. Specifying a number greater than zero enables text detection.
	MaxTexts int
	// DocumentText specifies whether a dense text document OCR should be run
	// on the image. When true, takes precedence over MaxTexts.
	DocumentText bool
	// SafeSearch specifies whether a safe-search detection should be run on the image.
	SafeSearch bool
	// ImageProps specifies whether image properties should be obtained for the image.
	ImageProps bool
	// Web specifies whether web annotations should be obtained for the image.
	Web bool
	// CropHints specifies whether crop hints should be computed for the image.
	CropHints *CropHintsParams
}

func (ar *AnnotateRequest) toProto() *pb.AnnotateImageRequest {
	img, ictx := ar.Image.toProtos()
	var features []*pb.Feature
	add := func(typ pb.Feature_Type, max int) {
		var mr int32
		if max > math.MaxInt32 {
			mr = math.MaxInt32
		} else {
			mr = int32(max)
		}
		features = append(features, &pb.Feature{Type: typ, MaxResults: mr})
	}
	if ar.MaxFaces > 0 {
		add(pb.Feature_FACE_DETECTION, ar.MaxFaces)
	}
	if ar.MaxLandmarks > 0 {
		add(pb.Feature_LANDMARK_DETECTION, ar.MaxLandmarks)
	}
	if ar.MaxLogos > 0 {
		add(pb.Feature_LOGO_DETECTION, ar.MaxLogos)
	}
	if ar.MaxLabels > 0 {
		add(pb.Feature_LABEL_DETECTION, ar.MaxLabels)
	}
	if ar.MaxTexts > 0 {
		add(pb.Feature_TEXT_DETECTION, ar.MaxTexts)
	}
	if ar.DocumentText {
		add(pb.Feature_DOCUMENT_TEXT_DETECTION, 0)
	}
	if ar.SafeSearch {
		add(pb.Feature_SAFE_SEARCH_DETECTION, 0)
	}
	if ar.ImageProps {
		add(pb.Feature_IMAGE_PROPERTIES, 0)
	}
	if ar.Web {
		add(pb.Feature_WEB_DETECTION, 0)
	}
	if ar.CropHints != nil {
		add(pb.Feature_CROP_HINTS, 0)
		if ictx == nil {
			ictx = &pb.ImageContext{}
		}
		ictx.CropHintsParams = &pb.CropHintsParams{
			AspectRatios: ar.CropHints.AspectRatios,
		}
	}
	return &pb.AnnotateImageRequest{
		Image:        img,
		Features:     features,
		ImageContext: ictx,
	}
}

// CropHintsParams are parameters for a request for crop hints.
type CropHintsParams struct {
	// Aspect ratios for desired crop hints, representing the ratio of the
	// width to the height of the image. For example, if the desired aspect
	// ratio is 4:3, the corresponding float value should be 1.33333. If not
	// specified, the best possible crop is returned. The number of provided
	// aspect ratios is limited to a maximum of 16; any aspect ratios provided
	// after the 16th are ignored.
	AspectRatios []float32
}

// Called for a single image and a single feature.
func (c *Client) annotateOne(ctx context.Context, req *AnnotateRequest) (*Annotations, error) {
	annsSlice, err := c.Annotate(ctx, req)
	if err != nil {
		return nil, err
	}
	anns := annsSlice[0]
	// When there is only one image and one feature, the Annotations.Error field is
	// unambiguously about that one detection, so we "promote" it to the error return value.
	if anns.Error != nil {
		return nil, anns.Error
	}
	return anns, nil
}

// TODO(jba): add examples for all single-feature functions (below).

// DetectFaces performs face detection on the image.
// At most maxResults results are returned.
func (c *Client) DetectFaces(ctx context.Context, img *Image, maxResults int) ([]*FaceAnnotation, error) {
	anns, err := c.annotateOne(ctx, &AnnotateRequest{Image: img, MaxFaces: maxResults})
	if err != nil {
		return nil, err
	}
	return anns.Faces, nil
}

// DetectLandmarks performs landmark detection on the image.
// At most maxResults results are returned.
func (c *Client) DetectLandmarks(ctx context.Context, img *Image, maxResults int) ([]*EntityAnnotation, error) {
	anns, err := c.annotateOne(ctx, &AnnotateRequest{Image: img, MaxLandmarks: maxResults})
	if err != nil {
		return nil, err
	}
	return anns.Landmarks, nil
}

// DetectLogos performs logo detection on the image.
// At most maxResults results are returned.
func (c *Client) DetectLogos(ctx context.Context, img *Image, maxResults int) ([]*EntityAnnotation, error) {
	anns, err := c.annotateOne(ctx, &AnnotateRequest{Image: img, MaxLogos: maxResults})
	if err != nil {
		return nil, err
	}
	return anns.Logos, nil
}

// DetectLabels performs label detection on the image.
// At most maxResults results are returned.
func (c *Client) DetectLabels(ctx context.Context, img *Image, maxResults int) ([]*EntityAnnotation, error) {
	anns, err := c.annotateOne(ctx, &AnnotateRequest{Image: img, MaxLabels: maxResults})
	if err != nil {
		return nil, err
	}
	return anns.Labels, nil
}

// DetectTexts performs text detection on the image.
// At most maxResults results are returned.
func (c *Client) DetectTexts(ctx context.Context, img *Image, maxResults int) ([]*EntityAnnotation, error) {
	anns, err := c.annotateOne(ctx, &AnnotateRequest{Image: img, MaxTexts: maxResults})
	if err != nil {
		return nil, err
	}
	return anns.Texts, nil
}

// DetectDocumentText performs full text (OCR) detection on the image.
func (c *Client) DetectDocumentText(ctx context.Context, img *Image) (*TextAnnotation, error) {
	anns, err := c.annotateOne(ctx, &AnnotateRequest{Image: img, DocumentText: true})
	if err != nil {
		return nil, err
	}
	return anns.FullText, nil
}

// DetectSafeSearch performs safe-search detection on the image.
func (c *Client) DetectSafeSearch(ctx context.Context, img *Image) (*SafeSearchAnnotation, error) {
	anns, err := c.annotateOne(ctx, &AnnotateRequest{Image: img, SafeSearch: true})
	if err != nil {
		return nil, err
	}
	return anns.SafeSearch, nil
}

// DetectImageProps computes properties of the image.
func (c *Client) DetectImageProps(ctx context.Context, img *Image) (*ImageProps, error) {
	anns, err := c.annotateOne(ctx, &AnnotateRequest{Image: img, ImageProps: true})
	if err != nil {
		return nil, err
	}
	return anns.ImageProps, nil
}

// DetectWeb computes a web annotation on the image.
func (c *Client) DetectWeb(ctx context.Context, img *Image) (*WebDetection, error) {
	anns, err := c.annotateOne(ctx, &AnnotateRequest{Image: img, Web: true})
	if err != nil {
		return nil, err
	}
	return anns.Web, nil
}

// CropHints computes crop hints for the image.
func (c *Client) CropHints(ctx context.Context, img *Image, params *CropHintsParams) ([]*CropHint, error) {
	// A nil AnnotateRequest.CropHints means do not perform CropHints. But
	// here the user is explicitly asking for CropHints, so treat nil as
	// an empty CropHintsParams.
	if params == nil {
		params = &CropHintsParams{}
	}
	anns, err := c.annotateOne(ctx, &AnnotateRequest{Image: img, CropHints: params})
	if err != nil {
		return nil, err
	}
	return anns.CropHints, nil
}

// A Likelihood is an approximate representation of a probability.
type Likelihood int

const (
	// LikelihoodUnknown means the likelihood is unknown.
	LikelihoodUnknown = Likelihood(pb.Likelihood_UNKNOWN)

	// VeryUnlikely means the image is very unlikely to belong to the feature specified.
	VeryUnlikely = Likelihood(pb.Likelihood_VERY_UNLIKELY)

	// Unlikely means the image is unlikely to belong to the feature specified.
	Unlikely = Likelihood(pb.Likelihood_UNLIKELY)

	// Possible means the image possibly belongs to the feature specified.
	Possible = Likelihood(pb.Likelihood_POSSIBLE)

	// Likely means the image is likely to belong to the feature specified.
	Likely = Likelihood(pb.Likelihood_LIKELY)

	// VeryLikely means the image is very likely to belong to the feature specified.
	VeryLikely = Likelihood(pb.Likelihood_VERY_LIKELY)
)

// A Property is an arbitrary name-value pair.
type Property struct {
	Name  string
	Value string
}

func propertyFromProto(p *pb.Property) Property {
	return Property{Name: p.Name, Value: p.Value}
}

// ColorInfo consists of RGB channels, score and fraction of
// image the color occupies in the image.
type ColorInfo struct {
	// RGB components of the color.
	Color color.NRGBA64

	// Score is the image-specific score for this color, in the range [0, 1].
	Score float32

	// PixelFraction is the fraction of pixels the color occupies in the image,
	// in the range [0, 1].
	PixelFraction float32
}

func colorInfoFromProto(ci *pb.ColorInfo) *ColorInfo {
	return &ColorInfo{
		Color:         colorFromProto(ci.Color),
		Score:         ci.Score,
		PixelFraction: ci.PixelFraction,
	}
}

// Should this go into protobuf/ptypes? The color proto is in google/types, so
// not specific to this API.
func colorFromProto(c *cpb.Color) color.NRGBA64 {
	// Convert a color component from [0.0, 1.0] to a uint16.
	cvt := func(f float32) uint16 { return uint16(f*math.MaxUint16 + 0.5) }

	var alpha float32 = 1
	if c.Alpha != nil {
		alpha = c.Alpha.Value
	}
	return color.NRGBA64{
		R: cvt(c.Red),
		G: cvt(c.Green),
		B: cvt(c.Blue),
		A: cvt(alpha),
	}
}
