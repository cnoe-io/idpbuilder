//go:build !ignore_autogenerated

/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgoCDPackageSpec) DeepCopyInto(out *ArgoCDPackageSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgoCDPackageSpec.
func (in *ArgoCDPackageSpec) DeepCopy() *ArgoCDPackageSpec {
	if in == nil {
		return nil
	}
	out := new(ArgoCDPackageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgoCDStatus) DeepCopyInto(out *ArgoCDStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgoCDStatus.
func (in *ArgoCDStatus) DeepCopy() *ArgoCDStatus {
	if in == nil {
		return nil
	}
	out := new(ArgoCDStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArgoPackageConfigSpec) DeepCopyInto(out *ArgoPackageConfigSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArgoPackageConfigSpec.
func (in *ArgoPackageConfigSpec) DeepCopy() *ArgoPackageConfigSpec {
	if in == nil {
		return nil
	}
	out := new(ArgoPackageConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BuildCustomizationSpec) DeepCopyInto(out *BuildCustomizationSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BuildCustomizationSpec.
func (in *BuildCustomizationSpec) DeepCopy() *BuildCustomizationSpec {
	if in == nil {
		return nil
	}
	out := new(BuildCustomizationSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Commit) DeepCopyInto(out *Commit) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Commit.
func (in *Commit) DeepCopy() *Commit {
	if in == nil {
		return nil
	}
	out := new(Commit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomPackage) DeepCopyInto(out *CustomPackage) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomPackage.
func (in *CustomPackage) DeepCopy() *CustomPackage {
	if in == nil {
		return nil
	}
	out := new(CustomPackage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomPackage) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomPackageList) DeepCopyInto(out *CustomPackageList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CustomPackage, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomPackageList.
func (in *CustomPackageList) DeepCopy() *CustomPackageList {
	if in == nil {
		return nil
	}
	out := new(CustomPackageList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CustomPackageList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomPackageSpec) DeepCopyInto(out *CustomPackageSpec) {
	*out = *in
	out.ArgoCD = in.ArgoCD
	out.GitServerAuthSecretRef = in.GitServerAuthSecretRef
	out.RemoteRepository = in.RemoteRepository
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomPackageSpec.
func (in *CustomPackageSpec) DeepCopy() *CustomPackageSpec {
	if in == nil {
		return nil
	}
	out := new(CustomPackageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomPackageStatus) DeepCopyInto(out *CustomPackageStatus) {
	*out = *in
	if in.GitRepositoryRefs != nil {
		in, out := &in.GitRepositoryRefs, &out.GitRepositoryRefs
		*out = make([]ObjectRef, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomPackageStatus.
func (in *CustomPackageStatus) DeepCopy() *CustomPackageStatus {
	if in == nil {
		return nil
	}
	out := new(CustomPackageStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EmbeddedArgoApplicationsPackageConfigSpec) DeepCopyInto(out *EmbeddedArgoApplicationsPackageConfigSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EmbeddedArgoApplicationsPackageConfigSpec.
func (in *EmbeddedArgoApplicationsPackageConfigSpec) DeepCopy() *EmbeddedArgoApplicationsPackageConfigSpec {
	if in == nil {
		return nil
	}
	out := new(EmbeddedArgoApplicationsPackageConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitRepository) DeepCopyInto(out *GitRepository) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitRepository.
func (in *GitRepository) DeepCopy() *GitRepository {
	if in == nil {
		return nil
	}
	out := new(GitRepository)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GitRepository) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitRepositoryList) DeepCopyInto(out *GitRepositoryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GitRepository, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitRepositoryList.
func (in *GitRepositoryList) DeepCopy() *GitRepositoryList {
	if in == nil {
		return nil
	}
	out := new(GitRepositoryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GitRepositoryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitRepositorySource) DeepCopyInto(out *GitRepositorySource) {
	*out = *in
	out.RemoteRepository = in.RemoteRepository
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitRepositorySource.
func (in *GitRepositorySource) DeepCopy() *GitRepositorySource {
	if in == nil {
		return nil
	}
	out := new(GitRepositorySource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitRepositorySpec) DeepCopyInto(out *GitRepositorySpec) {
	*out = *in
	out.Customization = in.Customization
	out.SecretRef = in.SecretRef
	out.Source = in.Source
	out.Provider = in.Provider
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitRepositorySpec.
func (in *GitRepositorySpec) DeepCopy() *GitRepositorySpec {
	if in == nil {
		return nil
	}
	out := new(GitRepositorySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitRepositoryStatus) DeepCopyInto(out *GitRepositoryStatus) {
	*out = *in
	out.LatestCommit = in.LatestCommit
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitRepositoryStatus.
func (in *GitRepositoryStatus) DeepCopy() *GitRepositoryStatus {
	if in == nil {
		return nil
	}
	out := new(GitRepositoryStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GiteaStatus) DeepCopyInto(out *GiteaStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GiteaStatus.
func (in *GiteaStatus) DeepCopy() *GiteaStatus {
	if in == nil {
		return nil
	}
	out := new(GiteaStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Localbuild) DeepCopyInto(out *Localbuild) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Localbuild.
func (in *Localbuild) DeepCopy() *Localbuild {
	if in == nil {
		return nil
	}
	out := new(Localbuild)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Localbuild) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocalbuildList) DeepCopyInto(out *LocalbuildList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Localbuild, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocalbuildList.
func (in *LocalbuildList) DeepCopy() *LocalbuildList {
	if in == nil {
		return nil
	}
	out := new(LocalbuildList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LocalbuildList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocalbuildSpec) DeepCopyInto(out *LocalbuildSpec) {
	*out = *in
	in.PackageConfigs.DeepCopyInto(&out.PackageConfigs)
	out.BuildCustomization = in.BuildCustomization
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocalbuildSpec.
func (in *LocalbuildSpec) DeepCopy() *LocalbuildSpec {
	if in == nil {
		return nil
	}
	out := new(LocalbuildSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocalbuildStatus) DeepCopyInto(out *LocalbuildStatus) {
	*out = *in
	out.ArgoCD = in.ArgoCD
	out.Nginx = in.Nginx
	out.Gitea = in.Gitea
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocalbuildStatus.
func (in *LocalbuildStatus) DeepCopy() *LocalbuildStatus {
	if in == nil {
		return nil
	}
	out := new(LocalbuildStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NginxStatus) DeepCopyInto(out *NginxStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NginxStatus.
func (in *NginxStatus) DeepCopy() *NginxStatus {
	if in == nil {
		return nil
	}
	out := new(NginxStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ObjectRef) DeepCopyInto(out *ObjectRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ObjectRef.
func (in *ObjectRef) DeepCopy() *ObjectRef {
	if in == nil {
		return nil
	}
	out := new(ObjectRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PackageConfigsSpec) DeepCopyInto(out *PackageConfigsSpec) {
	*out = *in
	out.Argo = in.Argo
	out.EmbeddedArgoApplications = in.EmbeddedArgoApplications
	if in.CustomPackageDirs != nil {
		in, out := &in.CustomPackageDirs, &out.CustomPackageDirs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.CustomPackageUrls != nil {
		in, out := &in.CustomPackageUrls, &out.CustomPackageUrls
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.CorePackageCustomization != nil {
		in, out := &in.CorePackageCustomization, &out.CorePackageCustomization
		*out = make(map[string]PackageCustomization, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PackageConfigsSpec.
func (in *PackageConfigsSpec) DeepCopy() *PackageConfigsSpec {
	if in == nil {
		return nil
	}
	out := new(PackageConfigsSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PackageCustomization) DeepCopyInto(out *PackageCustomization) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PackageCustomization.
func (in *PackageCustomization) DeepCopy() *PackageCustomization {
	if in == nil {
		return nil
	}
	out := new(PackageCustomization)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Provider) DeepCopyInto(out *Provider) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Provider.
func (in *Provider) DeepCopy() *Provider {
	if in == nil {
		return nil
	}
	out := new(Provider)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RemoteRepositorySpec) DeepCopyInto(out *RemoteRepositorySpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RemoteRepositorySpec.
func (in *RemoteRepositorySpec) DeepCopy() *RemoteRepositorySpec {
	if in == nil {
		return nil
	}
	out := new(RemoteRepositorySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretReference) DeepCopyInto(out *SecretReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretReference.
func (in *SecretReference) DeepCopy() *SecretReference {
	if in == nil {
		return nil
	}
	out := new(SecretReference)
	in.DeepCopyInto(out)
	return out
}
