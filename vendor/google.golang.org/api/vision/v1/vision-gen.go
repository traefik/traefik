// Package vision provides access to the Cloud Vision API.
//
// See https://cloud.google.com/vision/
//
// Usage example:
//
//   import "google.golang.org/api/vision/v1"
//   ...
//   visionService, err := vision.New(oauthHttpClient)
package vision // import "google.golang.org/api/vision/v1"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	context "golang.org/x/net/context"
	ctxhttp "golang.org/x/net/context/ctxhttp"
	gensupport "google.golang.org/api/gensupport"
	googleapi "google.golang.org/api/googleapi"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Always reference these packages, just in case the auto-generated code
// below doesn't.
var _ = bytes.NewBuffer
var _ = strconv.Itoa
var _ = fmt.Sprintf
var _ = json.NewDecoder
var _ = io.Copy
var _ = url.Parse
var _ = gensupport.MarshalJSON
var _ = googleapi.Version
var _ = errors.New
var _ = strings.Replace
var _ = context.Canceled
var _ = ctxhttp.Do

const apiId = "vision:v1"
const apiName = "vision"
const apiVersion = "v1"
const basePath = "https://vision.googleapis.com/"

// OAuth2 scopes used by this API.
const (
	// View and manage your data across Google Cloud Platform services
	CloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
)

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	s := &Service{client: client, BasePath: basePath}
	s.Images = NewImagesService(s)
	return s, nil
}

type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment

	Images *ImagesService
}

func (s *Service) userAgent() string {
	if s.UserAgent == "" {
		return googleapi.UserAgent
	}
	return googleapi.UserAgent + " " + s.UserAgent
}

func NewImagesService(s *Service) *ImagesService {
	rs := &ImagesService{s: s}
	return rs
}

type ImagesService struct {
	s *Service
}

// AnnotateImageRequest: Request for performing Vision tasks over a
// user-provided image, with
// user-requested features.
type AnnotateImageRequest struct {
	// Features: Requested features.
	Features []*Feature `json:"features,omitempty"`

	// Image: The image to be processed.
	Image *Image `json:"image,omitempty"`

	// ImageContext: Additional context that may accompany the image.
	ImageContext *ImageContext `json:"imageContext,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Features") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *AnnotateImageRequest) MarshalJSON() ([]byte, error) {
	type noMethod AnnotateImageRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// AnnotateImageResponse: Response to an image annotation request.
type AnnotateImageResponse struct {
	// Error: If set, represents the error message for the operation.
	// Note that filled-in mage annotations are guaranteed to be
	// correct, even when <code>error</code> is non-empty.
	Error *Status `json:"error,omitempty"`

	// FaceAnnotations: If present, face detection completed successfully.
	FaceAnnotations []*FaceAnnotation `json:"faceAnnotations,omitempty"`

	// ImagePropertiesAnnotation: If present, image properties were
	// extracted successfully.
	ImagePropertiesAnnotation *ImageProperties `json:"imagePropertiesAnnotation,omitempty"`

	// LabelAnnotations: If present, label detection completed successfully.
	LabelAnnotations []*EntityAnnotation `json:"labelAnnotations,omitempty"`

	// LandmarkAnnotations: If present, landmark detection completed
	// successfully.
	LandmarkAnnotations []*EntityAnnotation `json:"landmarkAnnotations,omitempty"`

	// LogoAnnotations: If present, logo detection completed successfully.
	LogoAnnotations []*EntityAnnotation `json:"logoAnnotations,omitempty"`

	// SafeSearchAnnotation: If present, safe-search annotation completed
	// successfully.
	SafeSearchAnnotation *SafeSearchAnnotation `json:"safeSearchAnnotation,omitempty"`

	// TextAnnotations: If present, text (OCR) detection completed
	// successfully.
	TextAnnotations []*EntityAnnotation `json:"textAnnotations,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Error") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *AnnotateImageResponse) MarshalJSON() ([]byte, error) {
	type noMethod AnnotateImageResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// BatchAnnotateImagesRequest: Multiple image annotation requests are
// batched into a single service call.
type BatchAnnotateImagesRequest struct {
	// Requests: Individual image annotation requests for this batch.
	Requests []*AnnotateImageRequest `json:"requests,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Requests") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *BatchAnnotateImagesRequest) MarshalJSON() ([]byte, error) {
	type noMethod BatchAnnotateImagesRequest
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// BatchAnnotateImagesResponse: Response to a batch image annotation
// request.
type BatchAnnotateImagesResponse struct {
	// Responses: Individual responses to image annotation requests within
	// the batch.
	Responses []*AnnotateImageResponse `json:"responses,omitempty"`

	// ServerResponse contains the HTTP response code and headers from the
	// server.
	googleapi.ServerResponse `json:"-"`

	// ForceSendFields is a list of field names (e.g. "Responses") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *BatchAnnotateImagesResponse) MarshalJSON() ([]byte, error) {
	type noMethod BatchAnnotateImagesResponse
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// BoundingPoly: A bounding polygon for the detected image annotation.
type BoundingPoly struct {
	// Vertices: The bounding polygon vertices.
	Vertices []*Vertex `json:"vertices,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Vertices") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *BoundingPoly) MarshalJSON() ([]byte, error) {
	type noMethod BoundingPoly
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Color: Represents a color in the RGBA color space. This
// representation is designed
// for simplicity of conversion to/from color representations in
// various
// languages over compactness; for example, the fields of this
// representation
// can be trivially provided to the constructor of "java.awt.Color" in
// Java; it
// can also be trivially provided to UIColor's
// "+colorWithRed:green:blue:alpha"
// method in iOS; and, with just a little work, it can be easily
// formatted into
// a CSS "rgba()" string in JavaScript, as well. Here are some
// examples:
//
// Example (Java):
//
//      import com.google.type.Color;
//
//      // ...
//      public static java.awt.Color fromProto(Color protocolor) {
//        float alpha = protocolor.hasAlpha()
//            ? protocolor.getAlpha().getValue()
//            : 1.0;
//
//        return new java.awt.Color(
//            protocolor.getRed(),
//            protocolor.getGreen(),
//            protocolor.getBlue(),
//            alpha);
//      }
//
//      public static Color toProto(java.awt.Color color) {
//        float red = (float) color.getRed();
//        float green = (float) color.getGreen();
//        float blue = (float) color.getBlue();
//        float denominator = 255.0;
//        Color.Builder resultBuilder =
//            Color
//                .newBuilder()
//                .setRed(red / denominator)
//                .setGreen(green / denominator)
//                .setBlue(blue / denominator);
//        int alpha = color.getAlpha();
//        if (alpha != 255) {
//          result.setAlpha(
//              FloatValue
//                  .newBuilder()
//                  .setValue(((float) alpha) / denominator)
//                  .build());
//        }
//        return resultBuilder.build();
//      }
//      // ...
//
// Example (iOS / Obj-C):
//
//      // ...
//      static UIColor* fromProto(Color* protocolor) {
//         float red = [protocolor red];
//         float green = [protocolor green];
//         float blue = [protocolor blue];
//         FloatValue* alpha_wrapper = [protocolor alpha];
//         float alpha = 1.0;
//         if (alpha_wrapper != nil) {
//           alpha = [alpha_wrapper value];
//         }
//         return [UIColor colorWithRed:red green:green blue:blue
// alpha:alpha];
//      }
//
//      static Color* toProto(UIColor* color) {
//          CGFloat red, green, blue, alpha;
//          if (![color getRed:&red green:&green blue:&blue
// alpha:&alpha]) {
//            return nil;
//          }
//          Color* result = [Color alloc] init];
//          [result setRed:red];
//          [result setGreen:green];
//          [result setBlue:blue];
//          if (alpha <= 0.9999) {
//            [result setAlpha:floatWrapperWithValue(alpha)];
//          }
//          [result autorelease];
//          return result;
//     }
//     // ...
//
//  Example (JavaScript):
//
//     // ...
//
//     var protoToCssColor = function(rgb_color) {
//        var redFrac = rgb_color.red || 0.0;
//        var greenFrac = rgb_color.green || 0.0;
//        var blueFrac = rgb_color.blue || 0.0;
//        var red = Math.floor(redFrac * 255);
//        var green = Math.floor(greenFrac * 255);
//        var blue = Math.floor(blueFrac * 255);
//
//        if (!('alpha' in rgb_color)) {
//           return rgbToCssColor_(red, green, blue);
//        }
//
//        var alphaFrac = rgb_color.alpha.value || 0.0;
//        var rgbParams = [red, green, blue].join(',');
//        return ['rgba(', rgbParams, ',', alphaFrac, ')'].join('');
//     };
//
//     var rgbToCssColor_ = function(red, green, blue) {
//       var rgbNumber = new Number((red << 16) | (green << 8) | blue);
//       var hexString = rgbNumber.toString(16);
//       var missingZeros = 6 - hexString.length;
//       var resultBuilder = ['#'];
//       for (var i = 0; i < missingZeros; i++) {
//          resultBuilder.push('0');
//       }
//       resultBuilder.push(hexString);
//       return resultBuilder.join('');
//     };
//
//     // ...
type Color struct {
	// Alpha: The fraction of this color that should be applied to the
	// pixel. That is,
	// the final pixel color is defined by the equation:
	//
	//   pixel color = alpha * (this color) + (1.0 - alpha) * (background
	// color)
	//
	// This means that a value of 1.0 corresponds to a solid color,
	// whereas
	// a value of 0.0 corresponds to a completely transparent color.
	// This
	// uses a wrapper message rather than a simple float scalar so that it
	// is
	// possible to distinguish between a default value and the value being
	// unset.
	// If omitted, this color object is to be rendered as a solid color
	// (as if the alpha value had been explicitly given with a value of
	// 1.0).
	Alpha float64 `json:"alpha,omitempty"`

	// Blue: The amount of blue in the color as a value in the interval [0,
	// 1].
	Blue float64 `json:"blue,omitempty"`

	// Green: The amount of green in the color as a value in the interval
	// [0, 1].
	Green float64 `json:"green,omitempty"`

	// Red: The amount of red in the color as a value in the interval [0,
	// 1].
	Red float64 `json:"red,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Alpha") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Color) MarshalJSON() ([]byte, error) {
	type noMethod Color
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ColorInfo: Color information consists of RGB channels, score and
// fraction of
// image the color occupies in the image.
type ColorInfo struct {
	// Color: RGB components of the color.
	Color *Color `json:"color,omitempty"`

	// PixelFraction: Stores the fraction of pixels the color occupies in
	// the image.
	// Value in range [0, 1].
	PixelFraction float64 `json:"pixelFraction,omitempty"`

	// Score: Image-specific score for this color. Value in range [0, 1].
	Score float64 `json:"score,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Color") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ColorInfo) MarshalJSON() ([]byte, error) {
	type noMethod ColorInfo
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// DominantColorsAnnotation: Set of dominant colors and their
// corresponding scores.
type DominantColorsAnnotation struct {
	// Colors: RGB color values, with their score and pixel fraction.
	Colors []*ColorInfo `json:"colors,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Colors") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *DominantColorsAnnotation) MarshalJSON() ([]byte, error) {
	type noMethod DominantColorsAnnotation
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// EntityAnnotation: Set of detected entity features.
type EntityAnnotation struct {
	// BoundingPoly: Image region to which this entity belongs.
	BoundingPoly *BoundingPoly `json:"boundingPoly,omitempty"`

	// Confidence: The accuracy of the entity detection in an image.
	// For example, for an image containing 'Eiffel Tower,' this field
	// represents
	// the confidence that there is a tower in the query image. Range [0,
	// 1].
	Confidence float64 `json:"confidence,omitempty"`

	// Description: Entity textual description, expressed in its
	// <code>locale</code> language.
	Description string `json:"description,omitempty"`

	// Locale: The language code for the locale in which the entity
	// textual
	// <code>description</code> (next field) is expressed.
	Locale string `json:"locale,omitempty"`

	// Locations: The location information for the detected entity.
	// Multiple
	// <code>LocationInfo</code> elements can be present since one location
	// may
	// indicate the location of the scene in the query image, and another
	// the
	// location of the place where the query image was taken. Location
	// information
	// is usually present for landmarks.
	Locations []*LocationInfo `json:"locations,omitempty"`

	// Mid: Knowledge Graph entity ID. Maps to a freebase entity ID.
	// (for example, "Google" maps to: mid /m/045c7b).
	Mid string `json:"mid,omitempty"`

	// Properties: Some entities can have additional optional
	// <code>Property</code> fields.
	// For example a different kind of score or string that qualifies the
	// entity.
	Properties []*Property `json:"properties,omitempty"`

	// Score: Overall score of the result. Range [0, 1].
	Score float64 `json:"score,omitempty"`

	// Topicality: The relevancy of the ICA (Image Content Annotation) label
	// to the
	// image. For example, the relevancy of 'tower' to an image
	// containing
	// 'Eiffel Tower' is likely higher than an image containing a distant
	// towering
	// building, though the confidence that there is a tower may be the
	// same.
	// Range [0, 1].
	Topicality float64 `json:"topicality,omitempty"`

	// ForceSendFields is a list of field names (e.g. "BoundingPoly") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *EntityAnnotation) MarshalJSON() ([]byte, error) {
	type noMethod EntityAnnotation
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// FaceAnnotation: A face annotation contains the results of face
// detection.
type FaceAnnotation struct {
	// AngerLikelihood: Anger likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - The image very unlikely belongs to the vertical
	// specified.
	//   "UNLIKELY" - The image unlikely belongs to the vertical specified.
	//   "POSSIBLE" - The image possibly belongs to the vertical specified.
	//   "LIKELY" - The image likely belongs to the vertical specified.
	//   "VERY_LIKELY" - The image very likely belongs to the vertical
	// specified.
	AngerLikelihood string `json:"angerLikelihood,omitempty"`

	// BlurredLikelihood: Blurred likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - The image very unlikely belongs to the vertical
	// specified.
	//   "UNLIKELY" - The image unlikely belongs to the vertical specified.
	//   "POSSIBLE" - The image possibly belongs to the vertical specified.
	//   "LIKELY" - The image likely belongs to the vertical specified.
	//   "VERY_LIKELY" - The image very likely belongs to the vertical
	// specified.
	BlurredLikelihood string `json:"blurredLikelihood,omitempty"`

	// BoundingPoly: The bounding polygon around the face. The coordinates
	// of the bounding box
	// are in the original image's scale, as returned in ImageParams.
	// The bounding box is computed to "frame" the face in accordance with
	// human
	// expectations. It is based on the landmarker results.
	// Note that one or more x and/or y coordinates may not be generated in
	// the
	// BoundingPoly (the polygon will be unbounded) if only a partial face
	// appears in
	// the image to be annotated.
	BoundingPoly *BoundingPoly `json:"boundingPoly,omitempty"`

	// DetectionConfidence: Detection confidence. Range [0, 1].
	DetectionConfidence float64 `json:"detectionConfidence,omitempty"`

	// FdBoundingPoly: This bounding polygon is tighter than the
	// previous
	// <code>boundingPoly</code>, and
	// encloses only the skin part of the face. Typically, it is used
	// to
	// eliminate the face from any image analysis that detects the
	// "amount of skin" visible in an image. It is not based on
	// the
	// landmarker results, only on the initial face detection, hence
	// the <code>fd</code> (face detection) prefix.
	FdBoundingPoly *BoundingPoly `json:"fdBoundingPoly,omitempty"`

	// HeadwearLikelihood: Headwear likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - The image very unlikely belongs to the vertical
	// specified.
	//   "UNLIKELY" - The image unlikely belongs to the vertical specified.
	//   "POSSIBLE" - The image possibly belongs to the vertical specified.
	//   "LIKELY" - The image likely belongs to the vertical specified.
	//   "VERY_LIKELY" - The image very likely belongs to the vertical
	// specified.
	HeadwearLikelihood string `json:"headwearLikelihood,omitempty"`

	// JoyLikelihood: Joy likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - The image very unlikely belongs to the vertical
	// specified.
	//   "UNLIKELY" - The image unlikely belongs to the vertical specified.
	//   "POSSIBLE" - The image possibly belongs to the vertical specified.
	//   "LIKELY" - The image likely belongs to the vertical specified.
	//   "VERY_LIKELY" - The image very likely belongs to the vertical
	// specified.
	JoyLikelihood string `json:"joyLikelihood,omitempty"`

	// LandmarkingConfidence: Face landmarking confidence. Range [0, 1].
	LandmarkingConfidence float64 `json:"landmarkingConfidence,omitempty"`

	// Landmarks: Detected face landmarks.
	Landmarks []*Landmark `json:"landmarks,omitempty"`

	// PanAngle: Yaw angle. Indicates the leftward/rightward angle that the
	// face is
	// pointing, relative to the vertical plane perpendicular to the image.
	// Range
	// [-180,180].
	PanAngle float64 `json:"panAngle,omitempty"`

	// RollAngle: Roll angle. Indicates the amount of
	// clockwise/anti-clockwise rotation of
	// the
	// face relative to the image vertical, about the axis perpendicular to
	// the
	// face. Range [-180,180].
	RollAngle float64 `json:"rollAngle,omitempty"`

	// SorrowLikelihood: Sorrow likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - The image very unlikely belongs to the vertical
	// specified.
	//   "UNLIKELY" - The image unlikely belongs to the vertical specified.
	//   "POSSIBLE" - The image possibly belongs to the vertical specified.
	//   "LIKELY" - The image likely belongs to the vertical specified.
	//   "VERY_LIKELY" - The image very likely belongs to the vertical
	// specified.
	SorrowLikelihood string `json:"sorrowLikelihood,omitempty"`

	// SurpriseLikelihood: Surprise likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - The image very unlikely belongs to the vertical
	// specified.
	//   "UNLIKELY" - The image unlikely belongs to the vertical specified.
	//   "POSSIBLE" - The image possibly belongs to the vertical specified.
	//   "LIKELY" - The image likely belongs to the vertical specified.
	//   "VERY_LIKELY" - The image very likely belongs to the vertical
	// specified.
	SurpriseLikelihood string `json:"surpriseLikelihood,omitempty"`

	// TiltAngle: Pitch angle. Indicates the upwards/downwards angle that
	// the face is
	// pointing
	// relative to the image's horizontal plane. Range [-180,180].
	TiltAngle float64 `json:"tiltAngle,omitempty"`

	// UnderExposedLikelihood: Under-exposed likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - The image very unlikely belongs to the vertical
	// specified.
	//   "UNLIKELY" - The image unlikely belongs to the vertical specified.
	//   "POSSIBLE" - The image possibly belongs to the vertical specified.
	//   "LIKELY" - The image likely belongs to the vertical specified.
	//   "VERY_LIKELY" - The image very likely belongs to the vertical
	// specified.
	UnderExposedLikelihood string `json:"underExposedLikelihood,omitempty"`

	// ForceSendFields is a list of field names (e.g. "AngerLikelihood") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *FaceAnnotation) MarshalJSON() ([]byte, error) {
	type noMethod FaceAnnotation
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Feature: The <em>Feature</em> indicates what type of image detection
// task to perform.
// Users describe the type of Vision tasks to perform over images
// by
// using <em>Feature</em>s. Features encode the Vision vertical to
// operate on
// and the number of top-scoring results to return.
type Feature struct {
	// MaxResults: Maximum number of results of this type.
	MaxResults int64 `json:"maxResults,omitempty"`

	// Type: The feature type.
	//
	// Possible values:
	//   "TYPE_UNSPECIFIED" - Unspecified feature type.
	//   "FACE_DETECTION" - Run face detection.
	//   "LANDMARK_DETECTION" - Run landmark detection.
	//   "LOGO_DETECTION" - Run logo detection.
	//   "LABEL_DETECTION" - Run label detection.
	//   "TEXT_DETECTION" - Run OCR.
	//   "SAFE_SEARCH_DETECTION" - Run various computer vision models to
	// compute image safe-search properties.
	//   "IMAGE_PROPERTIES" - Compute a set of properties about the image
	// (such as the image's dominant colors).
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "MaxResults") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Feature) MarshalJSON() ([]byte, error) {
	type noMethod Feature
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Image: Client image to perform Vision tasks over.
type Image struct {
	// Content: Image content, represented as a stream of bytes.
	Content string `json:"content,omitempty"`

	// Source: Google Cloud Storage image location. If both 'content' and
	// 'source'
	// are filled for an image, 'content' takes precedence and it will
	// be
	// used for performing the image annotation request.
	Source *ImageSource `json:"source,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Content") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Image) MarshalJSON() ([]byte, error) {
	type noMethod Image
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ImageContext: Image context.
type ImageContext struct {
	// LanguageHints: List of languages to use for TEXT_DETECTION. In most
	// cases, an empty value
	// will yield the best results as it will allow text detection
	// to
	// automatically detect the text language. For languages based on the
	// latin
	// alphabet a hint is not needed. In rare cases, when the language
	// of
	// the text in the image is known in advance, setting this hint will
	// help get
	// better results (although it will hurt a great deal if the hint is
	// wrong).
	// Text detection will return an error if one or more of the
	// languages
	// specified here are not supported. The exact list of supported
	// languages are
	// specified
	// here:
	// https://cloud.google.com/translate/v2/using_rest#language-params
	LanguageHints []string `json:"languageHints,omitempty"`

	// LatLongRect: Lat/long rectangle that specifies the location of the
	// image.
	LatLongRect *LatLongRect `json:"latLongRect,omitempty"`

	// ForceSendFields is a list of field names (e.g. "LanguageHints") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ImageContext) MarshalJSON() ([]byte, error) {
	type noMethod ImageContext
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ImageProperties: Stores image properties (e.g. dominant colors).
type ImageProperties struct {
	// DominantColors: If present, dominant colors completed successfully.
	DominantColors *DominantColorsAnnotation `json:"dominantColors,omitempty"`

	// ForceSendFields is a list of field names (e.g. "DominantColors") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ImageProperties) MarshalJSON() ([]byte, error) {
	type noMethod ImageProperties
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// ImageSource: External image source (i.e. Google Cloud Storage image
// location).
type ImageSource struct {
	// GcsImageUri: Google Cloud Storage image URI. It must be in the
	// following form:
	// "gs://bucket_name/object_name". For more
	// details, please see:
	// https://cloud.google.com/storage/docs/reference-uris.
	// NOTE: Cloud Storage object versioning is not supported!
	GcsImageUri string `json:"gcsImageUri,omitempty"`

	// ForceSendFields is a list of field names (e.g. "GcsImageUri") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *ImageSource) MarshalJSON() ([]byte, error) {
	type noMethod ImageSource
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Landmark: A face-specific landmark (for example, a face
// feature).
// Landmark positions may fall outside the bounds of the image
// when the face is near one or more edges of the image.
// Therefore it is NOT guaranteed that 0 <= x < width or 0 <= y <
// height.
type Landmark struct {
	// Position: Face landmark position.
	Position *Position `json:"position,omitempty"`

	// Type: Face landmark type.
	//
	// Possible values:
	//   "UNKNOWN_LANDMARK" - Unknown face landmark detected. Should not be
	// filled.
	//   "LEFT_EYE" - Left eye.
	//   "RIGHT_EYE" - Right eye.
	//   "LEFT_OF_LEFT_EYEBROW" - Left of left eyebrow.
	//   "RIGHT_OF_LEFT_EYEBROW" - Right of left eyebrow.
	//   "LEFT_OF_RIGHT_EYEBROW" - Left of right eyebrow.
	//   "RIGHT_OF_RIGHT_EYEBROW" - Right of right eyebrow.
	//   "MIDPOINT_BETWEEN_EYES" - Midpoint between eyes.
	//   "NOSE_TIP" - Nose tip.
	//   "UPPER_LIP" - Upper lip.
	//   "LOWER_LIP" - Lower lip.
	//   "MOUTH_LEFT" - Mouth left.
	//   "MOUTH_RIGHT" - Mouth right.
	//   "MOUTH_CENTER" - Mouth center.
	//   "NOSE_BOTTOM_RIGHT" - Nose, bottom right.
	//   "NOSE_BOTTOM_LEFT" - Nose, bottom left.
	//   "NOSE_BOTTOM_CENTER" - Nose, bottom center.
	//   "LEFT_EYE_TOP_BOUNDARY" - Left eye, top boundary.
	//   "LEFT_EYE_RIGHT_CORNER" - Left eye, right corner.
	//   "LEFT_EYE_BOTTOM_BOUNDARY" - Left eye, bottom boundary.
	//   "LEFT_EYE_LEFT_CORNER" - Left eye, left corner.
	//   "RIGHT_EYE_TOP_BOUNDARY" - Right eye, top boundary.
	//   "RIGHT_EYE_RIGHT_CORNER" - Right eye, right corner.
	//   "RIGHT_EYE_BOTTOM_BOUNDARY" - Right eye, bottom boundary.
	//   "RIGHT_EYE_LEFT_CORNER" - Right eye, left corner.
	//   "LEFT_EYEBROW_UPPER_MIDPOINT" - Left eyebrow, upper midpoint.
	//   "RIGHT_EYEBROW_UPPER_MIDPOINT" - Right eyebrow, upper midpoint.
	//   "LEFT_EAR_TRAGION" - Left ear tragion.
	//   "RIGHT_EAR_TRAGION" - Right ear tragion.
	//   "LEFT_EYE_PUPIL" - Left eye pupil.
	//   "RIGHT_EYE_PUPIL" - Right eye pupil.
	//   "FOREHEAD_GLABELLA" - Forehead glabella.
	//   "CHIN_GNATHION" - Chin gnathion.
	//   "CHIN_LEFT_GONION" - Chin left gonion.
	//   "CHIN_RIGHT_GONION" - Chin right gonion.
	Type string `json:"type,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Position") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Landmark) MarshalJSON() ([]byte, error) {
	type noMethod Landmark
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// LatLng: An object representing a latitude/longitude pair. This is
// expressed as a pair
// of doubles representing degrees latitude and degrees longitude.
// Unless
// specified otherwise, this must conform to the
// <a
// href="http://www.unoosa.org/pdf/icg/2012/template/WGS_84.pdf">WGS84
// st
// andard</a>. Values must be within normalized ranges.
//
// Example of normalization code in Python:
//
//     def NormalizeLongitude(longitude):
//       """Wraps decimal degrees longitude to [-180.0, 180.0]."""
//       q, r = divmod(longitude, 360.0)
//       if r > 180.0 or (r == 180.0 and q <= -1.0):
//         return r - 360.0
//       return r
//
//     def NormalizeLatLng(latitude, longitude):
//       """Wraps decimal degrees latitude and longitude to
//       [-180.0, 180.0] and [-90.0, 90.0], respectively."""
//       r = latitude % 360.0
//       if r <= 90.0:
//         return r, NormalizeLongitude(longitude)
//       elif r >= 270.0:
//         return r - 360, NormalizeLongitude(longitude)
//       else:
//         return 180 - r, NormalizeLongitude(longitude + 180.0)
//
//     assert 180.0 == NormalizeLongitude(180.0)
//     assert -180.0 == NormalizeLongitude(-180.0)
//     assert -179.0 == NormalizeLongitude(181.0)
//     assert (0.0, 0.0) == NormalizeLatLng(360.0, 0.0)
//     assert (0.0, 0.0) == NormalizeLatLng(-360.0, 0.0)
//     assert (85.0, 180.0) == NormalizeLatLng(95.0, 0.0)
//     assert (-85.0, -170.0) == NormalizeLatLng(-95.0, 10.0)
//     assert (90.0, 10.0) == NormalizeLatLng(90.0, 10.0)
//     assert (-90.0, -10.0) == NormalizeLatLng(-90.0, -10.0)
//     assert (0.0, -170.0) == NormalizeLatLng(-180.0, 10.0)
//     assert (0.0, -170.0) == NormalizeLatLng(180.0, 10.0)
//     assert (-90.0, 10.0) == NormalizeLatLng(270.0, 10.0)
//     assert (90.0, 10.0) == NormalizeLatLng(-270.0, 10.0)
type LatLng struct {
	// Latitude: The latitude in degrees. It must be in the range [-90.0,
	// +90.0].
	Latitude float64 `json:"latitude,omitempty"`

	// Longitude: The longitude in degrees. It must be in the range [-180.0,
	// +180.0].
	Longitude float64 `json:"longitude,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Latitude") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *LatLng) MarshalJSON() ([]byte, error) {
	type noMethod LatLng
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// LatLongRect: Rectangle determined by min and max LatLng pairs.
type LatLongRect struct {
	// MaxLatLng: Max lat/long pair.
	MaxLatLng *LatLng `json:"maxLatLng,omitempty"`

	// MinLatLng: Min lat/long pair.
	MinLatLng *LatLng `json:"minLatLng,omitempty"`

	// ForceSendFields is a list of field names (e.g. "MaxLatLng") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *LatLongRect) MarshalJSON() ([]byte, error) {
	type noMethod LatLongRect
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// LocationInfo: Detected entity location information.
type LocationInfo struct {
	// LatLng: Lat - long location coordinates.
	LatLng *LatLng `json:"latLng,omitempty"`

	// ForceSendFields is a list of field names (e.g. "LatLng") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *LocationInfo) MarshalJSON() ([]byte, error) {
	type noMethod LocationInfo
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Position: A 3D position in the image, used primarily for Face
// detection landmarks.
// A valid Position must have both x and y coordinates.
// The position coordinates are in the same scale as the original image.
type Position struct {
	// X: X coordinate.
	X float64 `json:"x,omitempty"`

	// Y: Y coordinate.
	Y float64 `json:"y,omitempty"`

	// Z: Z coordinate (or depth).
	Z float64 `json:"z,omitempty"`

	// ForceSendFields is a list of field names (e.g. "X") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Position) MarshalJSON() ([]byte, error) {
	type noMethod Position
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Property: Arbitrary name/value pair.
type Property struct {
	// Name: Name of the property.
	Name string `json:"name,omitempty"`

	// Value: Value of the property.
	Value string `json:"value,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Name") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Property) MarshalJSON() ([]byte, error) {
	type noMethod Property
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// SafeSearchAnnotation: Set of features pertaining to the image,
// computed by various computer vision
// methods over safe-search verticals (for example, adult, spoof,
// medical,
// violence).
type SafeSearchAnnotation struct {
	// Adult: Represents the adult contents likelihood for the image.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - The image very unlikely belongs to the vertical
	// specified.
	//   "UNLIKELY" - The image unlikely belongs to the vertical specified.
	//   "POSSIBLE" - The image possibly belongs to the vertical specified.
	//   "LIKELY" - The image likely belongs to the vertical specified.
	//   "VERY_LIKELY" - The image very likely belongs to the vertical
	// specified.
	Adult string `json:"adult,omitempty"`

	// Medical: Likelihood this is a medical image.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - The image very unlikely belongs to the vertical
	// specified.
	//   "UNLIKELY" - The image unlikely belongs to the vertical specified.
	//   "POSSIBLE" - The image possibly belongs to the vertical specified.
	//   "LIKELY" - The image likely belongs to the vertical specified.
	//   "VERY_LIKELY" - The image very likely belongs to the vertical
	// specified.
	Medical string `json:"medical,omitempty"`

	// Spoof: Spoof likelihood. The likelihood that an obvious
	// modification
	// was made to the image's canonical version to make it appear
	// funny or offensive.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - The image very unlikely belongs to the vertical
	// specified.
	//   "UNLIKELY" - The image unlikely belongs to the vertical specified.
	//   "POSSIBLE" - The image possibly belongs to the vertical specified.
	//   "LIKELY" - The image likely belongs to the vertical specified.
	//   "VERY_LIKELY" - The image very likely belongs to the vertical
	// specified.
	Spoof string `json:"spoof,omitempty"`

	// Violence: Violence likelihood.
	//
	// Possible values:
	//   "UNKNOWN" - Unknown likelihood.
	//   "VERY_UNLIKELY" - The image very unlikely belongs to the vertical
	// specified.
	//   "UNLIKELY" - The image unlikely belongs to the vertical specified.
	//   "POSSIBLE" - The image possibly belongs to the vertical specified.
	//   "LIKELY" - The image likely belongs to the vertical specified.
	//   "VERY_LIKELY" - The image very likely belongs to the vertical
	// specified.
	Violence string `json:"violence,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Adult") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *SafeSearchAnnotation) MarshalJSON() ([]byte, error) {
	type noMethod SafeSearchAnnotation
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// Status: The `Status` type defines a logical error model that is
// suitable for different
// programming environments, including REST APIs and RPC APIs. It is
// used by
// [gRPC](https://github.com/grpc). The error model is designed to
// be:
//
// - Simple to use and understand for most users
// - Flexible enough to meet unexpected needs
//
// # Overview
//
// The `Status` message contains three pieces of data: error code, error
// message,
// and error details. The error code should be an enum value
// of
// google.rpc.Code, but it may accept additional error codes if needed.
// The
// error message should be a developer-facing English message that
// helps
// developers *understand* and *resolve* the error. If a localized
// user-facing
// error message is needed, put the localized message in the error
// details or
// localize it in the client. The optional error details may contain
// arbitrary
// information about the error. There is a predefined set of error
// detail types
// in the package `google.rpc` which can be used for common error
// conditions.
//
// # Language mapping
//
// The `Status` message is the logical representation of the error
// model, but it
// is not necessarily the actual wire format. When the `Status` message
// is
// exposed in different client libraries and different wire protocols,
// it can be
// mapped differently. For example, it will likely be mapped to some
// exceptions
// in Java, but more likely mapped to some error codes in C.
//
// # Other uses
//
// The error model and the `Status` message can be used in a variety
// of
// environments, either with or without APIs, to provide a
// consistent developer experience across different
// environments.
//
// Example uses of this error model include:
//
// - Partial errors. If a service needs to return partial errors to the
// client,
//     it may embed the `Status` in the normal response to indicate the
// partial
//     errors.
//
// - Workflow errors. A typical workflow has multiple steps. Each step
// may
//     have a `Status` message for error reporting purpose.
//
// - Batch operations. If a client uses batch request and batch
// response, the
//     `Status` message should be used directly inside batch response,
// one for
//     each error sub-response.
//
// - Asynchronous operations. If an API call embeds asynchronous
// operation
//     results in its response, the status of those operations should
// be
//     represented directly using the `Status` message.
//
// - Logging. If some API errors are stored in logs, the message
// `Status` could
//     be used directly after any stripping needed for security/privacy
// reasons.
type Status struct {
	// Code: The status code, which should be an enum value of
	// google.rpc.Code.
	Code int64 `json:"code,omitempty"`

	// Details: A list of messages that carry the error details.  There will
	// be a
	// common set of message types for APIs to use.
	Details []StatusDetails `json:"details,omitempty"`

	// Message: A developer-facing error message, which should be in
	// English. Any
	// user-facing error message should be localized and sent in
	// the
	// google.rpc.Status.details field, or localized by the client.
	Message string `json:"message,omitempty"`

	// ForceSendFields is a list of field names (e.g. "Code") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Status) MarshalJSON() ([]byte, error) {
	type noMethod Status
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

type StatusDetails interface{}

// Vertex: A vertex represents a 2D point in the image.
// NOTE: the vertex coordinates are in the same scale as the original
// image.
type Vertex struct {
	// X: X coordinate.
	X int64 `json:"x,omitempty"`

	// Y: Y coordinate.
	Y int64 `json:"y,omitempty"`

	// ForceSendFields is a list of field names (e.g. "X") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	ForceSendFields []string `json:"-"`
}

func (s *Vertex) MarshalJSON() ([]byte, error) {
	type noMethod Vertex
	raw := noMethod(*s)
	return gensupport.MarshalJSON(raw, s.ForceSendFields)
}

// method id "vision.images.annotate":

type ImagesAnnotateCall struct {
	s                          *Service
	batchannotateimagesrequest *BatchAnnotateImagesRequest
	urlParams_                 gensupport.URLParams
	ctx_                       context.Context
}

// Annotate: Run image detection and annotation for a batch of images.
func (r *ImagesService) Annotate(batchannotateimagesrequest *BatchAnnotateImagesRequest) *ImagesAnnotateCall {
	c := &ImagesAnnotateCall{s: r.s, urlParams_: make(gensupport.URLParams)}
	c.batchannotateimagesrequest = batchannotateimagesrequest
	return c
}

// Fields allows partial responses to be retrieved. See
// https://developers.google.com/gdata/docs/2.0/basics#PartialResponse
// for more information.
func (c *ImagesAnnotateCall) Fields(s ...googleapi.Field) *ImagesAnnotateCall {
	c.urlParams_.Set("fields", googleapi.CombineFields(s))
	return c
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *ImagesAnnotateCall) Context(ctx context.Context) *ImagesAnnotateCall {
	c.ctx_ = ctx
	return c
}

func (c *ImagesAnnotateCall) doRequest(alt string) (*http.Response, error) {
	var body io.Reader = nil
	body, err := googleapi.WithoutDataWrapper.JSONReader(c.batchannotateimagesrequest)
	if err != nil {
		return nil, err
	}
	ctype := "application/json"
	c.urlParams_.Set("alt", alt)
	urls := googleapi.ResolveRelative(c.s.BasePath, "v1/images:annotate")
	urls += "?" + c.urlParams_.Encode()
	req, _ := http.NewRequest("POST", urls, body)
	googleapi.SetOpaque(req.URL)
	req.Header.Set("Content-Type", ctype)
	req.Header.Set("User-Agent", c.s.userAgent())
	if c.ctx_ != nil {
		return ctxhttp.Do(c.ctx_, c.s.client, req)
	}
	return c.s.client.Do(req)
}

// Do executes the "vision.images.annotate" call.
// Exactly one of *BatchAnnotateImagesResponse or error will be non-nil.
// Any non-2xx status code is an error. Response headers are in either
// *BatchAnnotateImagesResponse.ServerResponse.Header or (if a response
// was returned at all) in error.(*googleapi.Error).Header. Use
// googleapi.IsNotModified to check whether the returned error was
// because http.StatusNotModified was returned.
func (c *ImagesAnnotateCall) Do(opts ...googleapi.CallOption) (*BatchAnnotateImagesResponse, error) {
	gensupport.SetOptions(c.urlParams_, opts...)
	res, err := c.doRequest("json")
	if res != nil && res.StatusCode == http.StatusNotModified {
		if res.Body != nil {
			res.Body.Close()
		}
		return nil, &googleapi.Error{
			Code:   res.StatusCode,
			Header: res.Header,
		}
	}
	if err != nil {
		return nil, err
	}
	defer googleapi.CloseBody(res)
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	ret := &BatchAnnotateImagesResponse{
		ServerResponse: googleapi.ServerResponse{
			Header:         res.Header,
			HTTPStatusCode: res.StatusCode,
		},
	}
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
	// {
	//   "description": "Run image detection and annotation for a batch of images.",
	//   "flatPath": "v1/images:annotate",
	//   "httpMethod": "POST",
	//   "id": "vision.images.annotate",
	//   "parameterOrder": [],
	//   "parameters": {},
	//   "path": "v1/images:annotate",
	//   "request": {
	//     "$ref": "BatchAnnotateImagesRequest"
	//   },
	//   "response": {
	//     "$ref": "BatchAnnotateImagesResponse"
	//   },
	//   "scopes": [
	//     "https://www.googleapis.com/auth/cloud-platform"
	//   ]
	// }

}
