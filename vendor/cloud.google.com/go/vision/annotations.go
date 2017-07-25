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
	"image"

	"golang.org/x/text/language"
	pb "google.golang.org/genproto/googleapis/cloud/vision/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Annotations contains all the annotations performed by the API on a single image.
// A nil field indicates either that the corresponding feature was not requested,
// or that annotation failed for that feature.
type Annotations struct {
	// Faces holds the results of face detection.
	Faces []*FaceAnnotation
	// Landmarks holds the results of landmark detection.
	Landmarks []*EntityAnnotation
	// Logos holds the results of logo detection.
	Logos []*EntityAnnotation
	// Labels holds the results of label detection.
	Labels []*EntityAnnotation
	// Texts holds the results of text detection.
	Texts []*EntityAnnotation
	// FullText holds the results of full text (OCR) detection.
	FullText *TextAnnotation
	// SafeSearch holds the results of safe-search detection.
	SafeSearch *SafeSearchAnnotation
	// ImageProps contains properties of the annotated image.
	ImageProps *ImageProps
	// Web contains web annotations for the image.
	Web *WebDetection
	// CropHints contains crop hints for the image.
	CropHints []*CropHint

	// If non-nil, then one or more of the attempted annotations failed.
	// Non-nil annotations are guaranteed to be correct, even if Error is
	// non-nil.
	Error error
}

func annotationsFromProto(res *pb.AnnotateImageResponse) *Annotations {
	as := &Annotations{}
	for _, a := range res.FaceAnnotations {
		as.Faces = append(as.Faces, faceAnnotationFromProto(a))
	}
	for _, a := range res.LandmarkAnnotations {
		as.Landmarks = append(as.Landmarks, entityAnnotationFromProto(a))
	}
	for _, a := range res.LogoAnnotations {
		as.Logos = append(as.Logos, entityAnnotationFromProto(a))
	}
	for _, a := range res.LabelAnnotations {
		as.Labels = append(as.Labels, entityAnnotationFromProto(a))
	}
	for _, a := range res.TextAnnotations {
		as.Texts = append(as.Texts, entityAnnotationFromProto(a))
	}
	as.FullText = textAnnotationFromProto(res.FullTextAnnotation)
	as.SafeSearch = safeSearchAnnotationFromProto(res.SafeSearchAnnotation)
	as.ImageProps = imagePropertiesFromProto(res.ImagePropertiesAnnotation)
	as.Web = webDetectionFromProto(res.WebDetection)
	as.CropHints = cropHintsFromProto(res.CropHintsAnnotation)
	if res.Error != nil {
		// res.Error is a google.rpc.Status. Convert to a Go error. Use a gRPC
		// error because it preserves the code as a separate field.
		// TODO(jba): preserve the details field.
		as.Error = grpc.Errorf(codes.Code(res.Error.Code), "%s", res.Error.Message)
	}
	return as
}

// A FaceAnnotation describes the results of face detection on an image.
type FaceAnnotation struct {
	// BoundingPoly is the bounding polygon around the face. The coordinates of
	// the bounding box are in the original image's scale, as returned in
	// ImageParams. The bounding box is computed to "frame" the face in
	// accordance with human expectations. It is based on the landmarker
	// results. Note that one or more x and/or y coordinates may not be
	// generated in the BoundingPoly (the polygon will be unbounded) if only a
	// partial face appears in the image to be annotated.
	BoundingPoly []image.Point

	// FDBoundingPoly is tighter than BoundingPoly, and
	// encloses only the skin part of the face. Typically, it is used to
	// eliminate the face from any image analysis that detects the "amount of
	// skin" visible in an image. It is not based on the landmarker results, only
	// on the initial face detection, hence the fd (face detection) prefix.
	FDBoundingPoly []image.Point

	// Landmarks are detected face landmarks.
	Face FaceLandmarks

	// RollAngle indicates the amount of clockwise/anti-clockwise rotation of
	// the face relative to the image vertical, about the axis perpendicular to
	// the face. Range [-180,180].
	RollAngle float32

	// PanAngle is the yaw angle: the leftward/rightward angle that the face is
	// pointing, relative to the vertical plane perpendicular to the image. Range
	// [-180,180].
	PanAngle float32

	// TiltAngle is the pitch angle: the upwards/downwards angle that the face is
	// pointing relative to the image's horizontal plane. Range [-180,180].
	TiltAngle float32

	// DetectionConfidence is the detection confidence. The range is [0, 1].
	DetectionConfidence float32

	// LandmarkingConfidence is the face landmarking confidence. The range is [0, 1].
	LandmarkingConfidence float32

	// Likelihoods expresses the likelihood of various aspects of the face.
	Likelihoods *FaceLikelihoods
}

func faceAnnotationFromProto(pfa *pb.FaceAnnotation) *FaceAnnotation {
	fa := &FaceAnnotation{
		BoundingPoly:          boundingPolyFromProto(pfa.BoundingPoly),
		FDBoundingPoly:        boundingPolyFromProto(pfa.FdBoundingPoly),
		RollAngle:             pfa.RollAngle,
		PanAngle:              pfa.PanAngle,
		TiltAngle:             pfa.TiltAngle,
		DetectionConfidence:   pfa.DetectionConfidence,
		LandmarkingConfidence: pfa.LandmarkingConfidence,
		Likelihoods: &FaceLikelihoods{
			Joy:          Likelihood(pfa.JoyLikelihood),
			Sorrow:       Likelihood(pfa.SorrowLikelihood),
			Anger:        Likelihood(pfa.AngerLikelihood),
			Surprise:     Likelihood(pfa.SurpriseLikelihood),
			UnderExposed: Likelihood(pfa.UnderExposedLikelihood),
			Blurred:      Likelihood(pfa.BlurredLikelihood),
			Headwear:     Likelihood(pfa.HeadwearLikelihood),
		},
	}
	populateFaceLandmarks(pfa.Landmarks, &fa.Face)
	return fa
}

// An EntityAnnotation describes the results of a landmark, label, logo or text
// detection on an image.
type EntityAnnotation struct {
	// ID is an opaque entity ID. Some IDs might be available in Knowledge Graph(KG).
	// For more details on KG please see:
	// https://developers.google.com/knowledge-graph/
	ID string

	// Locale is the language code for the locale in which the entity textual
	// description (next field) is expressed.
	Locale string

	// Description is the entity textual description, expressed in the language of Locale.
	Description string

	// Score is the overall score of the result. Range [0, 1].
	Score float32

	// Confidence is the accuracy of the entity detection in an image.
	// For example, for an image containing the Eiffel Tower, this field represents
	// the confidence that there is a tower in the query image. Range [0, 1].
	Confidence float32

	// Topicality is the relevancy of the ICA (Image Content Annotation) label to the
	// image. For example, the relevancy of 'tower' to an image containing
	// 'Eiffel Tower' is likely higher than an image containing a distant towering
	// building, though the confidence that there is a tower may be the same.
	// Range [0, 1].
	Topicality float32

	// BoundingPoly is the image region to which this entity belongs. Not filled currently
	// for label detection. For text detection, BoundingPolys
	// are produced for the entire text detected in an image region, followed by
	// BoundingPolys for each word within the detected text.
	BoundingPoly []image.Point

	// Locations contains the location information for the detected entity.
	// Multiple LatLng structs can be present since one location may indicate the
	// location of the scene in the query image, and another the location of the
	// place where the query image was taken. Location information is usually
	// present for landmarks.
	Locations []LatLng

	// Properties are additional optional Property fields.
	// For example a different kind of score or string that qualifies the entity.
	Properties []Property
}

func entityAnnotationFromProto(e *pb.EntityAnnotation) *EntityAnnotation {
	var locs []LatLng
	for _, li := range e.Locations {
		locs = append(locs, latLngFromProto(li.LatLng))
	}
	var props []Property
	for _, p := range e.Properties {
		props = append(props, propertyFromProto(p))
	}
	return &EntityAnnotation{
		ID:           e.Mid,
		Locale:       e.Locale,
		Description:  e.Description,
		Score:        e.Score,
		Confidence:   e.Confidence,
		Topicality:   e.Topicality,
		BoundingPoly: boundingPolyFromProto(e.BoundingPoly),
		Locations:    locs,
		Properties:   props,
	}
}

// TextAnnotation contains a structured representation of OCR extracted text.
// The hierarchy of an OCR extracted text structure looks like:
//     TextAnnotation -> Page -> Block -> Paragraph -> Word -> Symbol
// Each structural component, starting from Page, may further have its own
// properties. Properties describe detected languages, breaks etc.
type TextAnnotation struct {
	// List of pages detected by OCR.
	Pages []*Page
	// UTF-8 text detected on the pages.
	Text string
}

func textAnnotationFromProto(pta *pb.TextAnnotation) *TextAnnotation {
	if pta == nil {
		return nil
	}
	var pages []*Page
	for _, p := range pta.Pages {
		pages = append(pages, pageFromProto(p))
	}
	return &TextAnnotation{
		Pages: pages,
		Text:  pta.Text,
	}
}

// A Page is a page of text detected from OCR.
type Page struct {
	// Additional information detected on the page.
	Properties *TextProperties
	// Page width in pixels.
	Width int32
	// Page height in pixels.
	Height int32
	// List of blocks of text, images etc on this page.
	Blocks []*Block
}

func pageFromProto(p *pb.Page) *Page {
	if p == nil {
		return nil
	}
	var blocks []*Block
	for _, b := range p.Blocks {
		blocks = append(blocks, blockFromProto(b))
	}
	return &Page{
		Properties: textPropertiesFromProto(p.Property),
		Width:      p.Width,
		Height:     p.Height,
		Blocks:     blocks,
	}
}

// A Block is a logical element on the page.
type Block struct {
	// Additional information detected for the block.
	Properties *TextProperties
	// The bounding box for the block.
	// The vertices are in the order of top-left, top-right, bottom-right,
	// bottom-left. When a rotation of the bounding box is detected the rotation
	// is represented as around the top-left corner as defined when the text is
	// read in the 'natural' orientation.
	// For example:
	//   * when the text is horizontal it might look like:
	//      0----1
	//      |    |
	//      3----2
	//   * when it's rotated 180 degrees around the top-left corner it becomes:
	//      2----3
	//      |    |
	//      1----0
	//   and the vertice order will still be (0, 1, 2, 3).
	BoundingBox []image.Point
	// List of paragraphs in this block (if this blocks is of type text).
	Paragraphs []*Paragraph
	// Detected block type (text, image etc) for this block.
	BlockType BlockType
}

// A BlockType represents the kind of Block (text, image, etc.)
type BlockType int

const (
	// Unknown block type.
	UnknownBlock BlockType = BlockType(pb.Block_UNKNOWN)
	// Regular text block.
	TextBlock BlockType = BlockType(pb.Block_TEXT)
	// Table block.
	TableBlock BlockType = BlockType(pb.Block_TABLE)
	// Image block.
	PictureBlock BlockType = BlockType(pb.Block_PICTURE)
	// Horizontal/vertical line box.
	RulerBlock BlockType = BlockType(pb.Block_RULER)
	// Barcode block.
	BarcodeBlock BlockType = BlockType(pb.Block_BARCODE)
)

func blockFromProto(p *pb.Block) *Block {
	if p == nil {
		return nil
	}
	var paras []*Paragraph
	for _, pa := range p.Paragraphs {
		paras = append(paras, paragraphFromProto(pa))
	}
	return &Block{
		Properties:  textPropertiesFromProto(p.Property),
		BoundingBox: boundingPolyFromProto(p.BoundingBox),
		Paragraphs:  paras,
		BlockType:   BlockType(p.BlockType),
	}
}

// A Paragraph is a structural unit of text representing a number of words in
// certain order.
type Paragraph struct {
	// Additional information detected for the paragraph.
	Properties *TextProperties
	// The bounding box for the paragraph.
	// The vertices are in the order of top-left, top-right, bottom-right,
	// bottom-left. When a rotation of the bounding box is detected the rotation
	// is represented as around the top-left corner as defined when the text is
	// read in the 'natural' orientation.
	// For example:
	//   * when the text is horizontal it might look like:
	//      0----1
	//      |    |
	//      3----2
	//   * when it's rotated 180 degrees around the top-left corner it becomes:
	//      2----3
	//      |    |
	//      1----0
	//   and the vertice order will still be (0, 1, 2, 3).
	BoundingBox []image.Point
	// List of words in this paragraph.
	Words []*Word
}

func paragraphFromProto(p *pb.Paragraph) *Paragraph {
	if p == nil {
		return nil
	}
	var words []*Word
	for _, w := range p.Words {
		words = append(words, wordFromProto(w))
	}
	return &Paragraph{
		Properties:  textPropertiesFromProto(p.Property),
		BoundingBox: boundingPolyFromProto(p.BoundingBox),
		Words:       words,
	}
}

// A Word is a word in a text document.
type Word struct {
	// Additional information detected for the word.
	Properties *TextProperties
	// The bounding box for the word.
	// The vertices are in the order of top-left, top-right, bottom-right,
	// bottom-left. When a rotation of the bounding box is detected the rotation
	// is represented as around the top-left corner as defined when the text is
	// read in the 'natural' orientation.
	// For example:
	//   * when the text is horizontal it might look like:
	//      0----1
	//      |    |
	//      3----2
	//   * when it's rotated 180 degrees around the top-left corner it becomes:
	//      2----3
	//      |    |
	//      1----0
	//   and the vertice order will still be (0, 1, 2, 3).
	BoundingBox []image.Point
	// List of symbols in the word.
	// The order of the symbols follows the natural reading order.
	Symbols []*Symbol
}

func wordFromProto(p *pb.Word) *Word {
	if p == nil {
		return nil
	}
	var syms []*Symbol
	for _, s := range p.Symbols {
		syms = append(syms, symbolFromProto(s))
	}
	return &Word{
		Properties:  textPropertiesFromProto(p.Property),
		BoundingBox: boundingPolyFromProto(p.BoundingBox),
		Symbols:     syms,
	}
}

// A Symbol is a symbol in a text document.
type Symbol struct {
	// Additional information detected for the symbol.
	Properties *TextProperties
	// The bounding box for the symbol.
	// The vertices are in the order of top-left, top-right, bottom-right,
	// bottom-left. When a rotation of the bounding box is detected the rotation
	// is represented as around the top-left corner as defined when the text is
	// read in the 'natural' orientation.
	// For example:
	//   * when the text is horizontal it might look like:
	//      0----1
	//      |    |
	//      3----2
	//   * when it's rotated 180 degrees around the top-left corner it becomes:
	//      2----3
	//      |    |
	//      1----0
	//   and the vertice order will still be (0, 1, 2, 3).
	BoundingBox []image.Point
	// The actual UTF-8 representation of the symbol.
	Text string
}

func symbolFromProto(p *pb.Symbol) *Symbol {
	if p == nil {
		return nil
	}
	return &Symbol{
		Properties:  textPropertiesFromProto(p.Property),
		BoundingBox: boundingPolyFromProto(p.BoundingBox),
		Text:        p.Text,
	}
}

// TextProperties contains additional information about an OCR structural component.
type TextProperties struct {
	// A list of detected languages together with confidence.
	DetectedLanguages []*DetectedLanguage
	// Detected start or end of a text segment.
	DetectedBreak *DetectedBreak
}

// Detected language for a structural component.
type DetectedLanguage struct {
	// The BCP-47 language code, such as "en-US" or "sr-Latn".
	Code language.Tag
	// The confidence of the detected language, in the range [0, 1].
	Confidence float32
}

// DetectedBreak is the detected start or end of a structural component.
type DetectedBreak struct {
	// The type of break.
	Type DetectedBreakType
	// True if break prepends the element.
	IsPrefix bool
}

type DetectedBreakType int

const (
	// Unknown break label type.
	UnknownBreak = DetectedBreakType(pb.TextAnnotation_DetectedBreak_UNKNOWN)
	// Regular space.
	SpaceBreak = DetectedBreakType(pb.TextAnnotation_DetectedBreak_SPACE)
	// Sure space (very wide).
	SureSpaceBreak = DetectedBreakType(pb.TextAnnotation_DetectedBreak_SURE_SPACE)
	// Line-wrapping break.
	EOLSureSpaceBreak = DetectedBreakType(pb.TextAnnotation_DetectedBreak_EOL_SURE_SPACE)
	// End-line hyphen that is not present in text; does not co-occur with SPACE, LEADER_SPACE, or LINE_BREAK.
	HyphenBreak = DetectedBreakType(pb.TextAnnotation_DetectedBreak_HYPHEN)
	// Line break that ends a paragraph.
	LineBreak = DetectedBreakType(pb.TextAnnotation_DetectedBreak_LINE_BREAK)
)

func textPropertiesFromProto(p *pb.TextAnnotation_TextProperty) *TextProperties {
	var dls []*DetectedLanguage
	for _, dl := range p.DetectedLanguages {
		tag, _ := language.Parse(dl.LanguageCode)
		// Ignore error. If err != nil the returned tag will not be garbage,
		// but a best-effort attempt at a parse. At worst it will be
		// language.Und, the documented "undefined" Tag.
		dls = append(dls, &DetectedLanguage{Code: tag, Confidence: dl.Confidence})
	}
	var db *DetectedBreak
	if p.DetectedBreak != nil {
		db = &DetectedBreak{
			Type:     DetectedBreakType(p.DetectedBreak.Type),
			IsPrefix: p.DetectedBreak.IsPrefix,
		}
	}
	return &TextProperties{
		DetectedLanguages: dls,
		DetectedBreak:     db,
	}
}

// SafeSearchAnnotation describes the results of a SafeSearch detection on an image.
type SafeSearchAnnotation struct {
	// Adult is the likelihood that the image contains adult content.
	Adult Likelihood

	// Spoof is the likelihood that an obvious modification was made to the
	// image's canonical version to make it appear funny or offensive.
	Spoof Likelihood

	// Medical is the likelihood that this is a medical image.
	Medical Likelihood

	// Violence is the likelihood that this image represents violence.
	Violence Likelihood
}

func safeSearchAnnotationFromProto(s *pb.SafeSearchAnnotation) *SafeSearchAnnotation {
	if s == nil {
		return nil
	}
	return &SafeSearchAnnotation{
		Adult:    Likelihood(s.Adult),
		Spoof:    Likelihood(s.Spoof),
		Medical:  Likelihood(s.Medical),
		Violence: Likelihood(s.Violence),
	}
}

// ImageProps describes properties of the image itself, like the dominant colors.
type ImageProps struct {
	// DominantColors describes the dominant colors of the image.
	DominantColors []*ColorInfo
}

func imagePropertiesFromProto(ip *pb.ImageProperties) *ImageProps {
	if ip == nil || ip.DominantColors == nil {
		return nil
	}
	var cinfos []*ColorInfo
	for _, ci := range ip.DominantColors.Colors {
		cinfos = append(cinfos, colorInfoFromProto(ci))
	}
	return &ImageProps{DominantColors: cinfos}
}

// WebDetection contains relevant information for the image from the Internet.
type WebDetection struct {
	// Deduced entities from similar images on the Internet.
	WebEntities []*WebEntity
	// Fully matching images from the Internet.
	// They're definite neardups and most often a copy of the query image with
	// merely a size change.
	FullMatchingImages []*WebImage
	// Partial matching images from the Internet.
	// Those images are similar enough to share some key-point features. For
	// example an original image will likely have partial matching for its crops.
	PartialMatchingImages []*WebImage
	// Web pages containing the matching images from the Internet.
	PagesWithMatchingImages []*WebPage
}

func webDetectionFromProto(p *pb.WebDetection) *WebDetection {
	if p == nil {
		return nil
	}
	var (
		wes        []*WebEntity
		fmis, pmis []*WebImage
		wps        []*WebPage
	)
	for _, e := range p.WebEntities {
		wes = append(wes, webEntityFromProto(e))
	}
	for _, m := range p.FullMatchingImages {
		fmis = append(fmis, webImageFromProto(m))
	}
	for _, m := range p.PartialMatchingImages {
		pmis = append(fmis, webImageFromProto(m))
	}
	for _, g := range p.PagesWithMatchingImages {
		wps = append(wps, webPageFromProto(g))
	}
	return &WebDetection{
		WebEntities:             wes,
		FullMatchingImages:      fmis,
		PartialMatchingImages:   pmis,
		PagesWithMatchingImages: wps,
	}
}

// A WebEntity is an entity deduced from similar images on the Internet.
type WebEntity struct {
	// Opaque entity ID.
	ID string
	// Overall relevancy score for the entity.
	// Not normalized and not comparable across different image queries.
	Score float32
	// Canonical description of the entity, in English.
	Description string
}

func webEntityFromProto(p *pb.WebDetection_WebEntity) *WebEntity {
	return &WebEntity{
		ID:          p.EntityId,
		Score:       p.Score,
		Description: p.Description,
	}
}

// WebImage contains metadata for online images.
type WebImage struct {
	// The result image URL.
	URL string
	// Overall relevancy score for the image.
	// Not normalized and not comparable across different image queries.
	Score float32
}

func webImageFromProto(p *pb.WebDetection_WebImage) *WebImage {
	return &WebImage{
		URL:   p.Url,
		Score: p.Score,
	}
}

// A WebPage contains metadata for web pages.
type WebPage struct {
	// The result web page URL.
	URL string
	// Overall relevancy score for the web page.
	// Not normalized and not comparable across different image queries.
	Score float32
}

func webPageFromProto(p *pb.WebDetection_WebPage) *WebPage {
	return &WebPage{
		URL:   p.Url,
		Score: p.Score,
	}
}

// CropHint is a single crop hint that is used to generate a new crop when
// serving an image.
type CropHint struct {
	// The bounding polygon for the crop region. The coordinates of the bounding
	// box are in the original image's scale, as returned in `ImageParams`.
	BoundingPoly []image.Point
	// Confidence of this being a salient region.  Range [0, 1].
	Confidence float32
	// Fraction of importance of this salient region with respect to the original
	// image.
	ImportanceFraction float32
}

func cropHintsFromProto(p *pb.CropHintsAnnotation) []*CropHint {
	if p == nil {
		return nil
	}
	var chs []*CropHint
	for _, pch := range p.CropHints {
		chs = append(chs, cropHintFromProto(pch))
	}
	return chs
}

func cropHintFromProto(pch *pb.CropHint) *CropHint {
	return &CropHint{
		BoundingPoly:       boundingPolyFromProto(pch.BoundingPoly),
		Confidence:         pch.Confidence,
		ImportanceFraction: pch.ImportanceFraction,
	}
}
