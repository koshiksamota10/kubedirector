// Copyright 2020 Hewlett Packard Enterprise Development LP

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/common/log"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func convert(
	cr *v1beta1.ConversionReview,
) *v1beta1.ConversionResponse {

	var convertedObjects []runtime.RawExtension
	var conversionResponse = v1beta1.ConversionResponse{
		ConvertedObjects: convertedObjects,
		Result: metav1.Status{
			Message: metav1.StatusSuccess,
		},
	}

	//log.Info("Inside convert function ", ",desired version: ", cr)

	return &conversionResponse
}

func convertor(
	w http.ResponseWriter,
	r *http.Request,
) {
	//log.Info("I am in convertor, Mission Success! More work to do")
	var conversionResponse *v1beta1.ConversionResponse

	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	log.Info("Header: ", r.Header)
	if contentType != "application/json" {
		http.Error(
			w,
			"invalid Content-Type, expect `application/json`",
			http.StatusUnsupportedMediaType,
		)
		return
	}

	cr := v1beta1.ConversionReview{}
	if err := json.Unmarshal(body, &cr); err != nil {
		conversionResponse = &v1beta1.ConversionResponse{
			Result: metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		log.Info("Conversion Request:", cr.Request)
		conversionResponse = convert(&cr)
	}

	cr.Request = &v1beta1.ConversionRequest{}
	//log.Info("I am in convertor, Conversion Response has anything?!", conversionResponse)

	conversionReview := v1beta1.ConversionReview{}
	//log.Info("Conversion Response: ", conversionResponse)
	if conversionResponse != nil {
		conversionReview.Response = conversionResponse
		if cr.Request != nil {
			conversionReview.Response.UID = cr.Request.UID
		}
	}

	respBytes, err := json.Marshal(conversionReview)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("could not encode response: %v", err),
			http.StatusInternalServerError,
		)
	}
	if _, err := w.Write(respBytes); err != nil {
		http.Error(
			w,
			fmt.Sprintf("could not write response: %v", err),
			http.StatusInternalServerError,
		)
	}

}
