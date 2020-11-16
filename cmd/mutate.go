// Package mutate deals with AdmissionReview requests and responses, it takes in the request body and returns a readily converted JSON []byte that can be
// returned from a http Handler w/o needing to further convert or modify it, it also makes testing Mutate() kind of easy w/o need for a fake http server, etc.
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"gomodules.xyz/jsonpatch/v3"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Mutate Reads an admission review request and mutates it
func Mutate(body []byte) ([]byte, error) {
	// unmarshal request into AdmissionReview struct
	admReview := admissionv1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &admReview); err != nil {
		return nil, fmt.Errorf("unmarshaling request failed with %s", err)
	}

	var Job batchv1.Job
	ar := admReview.Request
	// get the Job object and unmarshal it into its struct, if we cannot, we might as well stop here
	if err := json.Unmarshal(ar.Object.Raw, &Job); err != nil {
		return nil, fmt.Errorf("unable unmarshal pod json object %v", err)
	}

	annotations := Job.Spec.Template.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations["linkerd.io/inject"] = "disabled"
	var admReviewResult *admissionv1beta1.AdmissionReview

	mutatedJSON, err := json.Marshal(Job)
	if err != nil {
		log.Printf("%s/%s failed mutation, failed to marshal Job, error: %s", Job.Namespace, Job.Name, err.Error())
		admReviewResult = failedAdmissionReview(ar.UID)
	}

	patch, err := jsonpatch.CreatePatch(ar.Object.Raw, mutatedJSON)
	if err != nil {
		log.Printf("%s/%s failed mutation, failed to CreatePath, error: %s", Job.Namespace, Job.Name, err.Error())
		admReviewResult = failedAdmissionReview(ar.UID)
	}

	marshalledPatch, err := json.Marshal(patch)
	if err != nil {
		log.Printf("%s/%s failed mutation, failed to marshal patch, error: %s", Job.Namespace, Job.Name, err.Error())
		admReviewResult = failedAdmissionReview(ar.UID)
	}
	log.Printf("%s/%s mutated", Job.Namespace, Job.Name)
	if admReviewResult == nil {
		admReviewResult = &admissionv1beta1.AdmissionReview{
			Response: &admissionv1beta1.AdmissionResponse{
				UID:       ar.UID,
				Allowed:   true,
				Patch:     marshalledPatch,
				PatchType: jsonPatchType,
				Result: &metav1.Status{
					Status: "Success",
				},
			},
		}
	} else {
		mutationErrors.Inc()
	}

	// back into JSON so we can return the finished AdmissionReview w/ Response directly
	// w/o needing to convert things in the http handler
	responseBody, err := json.Marshal(admReviewResult)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func failedAdmissionReview(uid types.UID) *admissionv1beta1.AdmissionReview {
	return &admissionv1beta1.AdmissionReview{
		Response: &admissionv1beta1.AdmissionResponse{
			UID:     uid,
			Allowed: false,
			Result: &metav1.Status{
				Status: "Failure",
			},
		},
	}
}

// jsonPatchType is the type for Kubernetes responses type.
var jsonPatchType = func() *admissionv1beta1.PatchType {
	pt := admissionv1beta1.PatchTypeJSONPatch
	return &pt
}()
