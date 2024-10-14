package main

import (
	admissionv1 "k8s.io/api/admission/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func toV1AdmissionErrorResp(err error) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
			Status:  metav1.StatusFailure,
		},
	}
}

func convertV1AdmitFuncToV1beta1(f admitv1Func) admitv1veta1Func {
	// 这里的转换需要入参类型和返回类型都转换
	// 对于v1beta的AdmissionReview，我们把它转换为v1的AdmissionReview
	// 然后转换即可
	return func(review admissionv1beta1.AdmissionReview) *admissionv1beta1.AdmissionResponse {
		in := admissionv1.AdmissionReview{Request: convertAdmissionRequestToV1(review.Request)}
		out := f(in)
		return convertAdmissionResponseToV1beta1(out)
	}
}

func convertAdmissionResponseToV1beta1(out *admissionv1.AdmissionResponse) *admissionv1beta1.AdmissionResponse {
	return &admissionv1beta1.AdmissionResponse{
		UID:              out.UID,
		Allowed:          out.Allowed,
		AuditAnnotations: out.AuditAnnotations,
		Patch:            out.Patch,
		PatchType:        (*admissionv1beta1.PatchType)(out.PatchType),
		Result:           out.Result,
		Warnings:         out.Warnings,
	}
}

func convertAdmissionRequestToV1(admissionRequest *admissionv1beta1.AdmissionRequest) *admissionv1.AdmissionRequest {
	return &admissionv1.AdmissionRequest{
		UID:         admissionRequest.UID,
		Kind:        admissionRequest.Kind,
		Resource:    admissionRequest.Resource,
		SubResource: admissionRequest.SubResource,
		Name:        admissionRequest.Name,
		Namespace:   admissionRequest.Namespace,
		Operation:   admissionv1.Operation(admissionRequest.Operation),
		UserInfo:    admissionRequest.UserInfo,
		Object:      admissionRequest.Object,
		OldObject:   admissionRequest.OldObject,
		DryRun:      admissionRequest.DryRun,
	}
}
