//  Copyright 2020 Google Inc. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package config

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/osconfig/packages"
	utilmocks "github.com/GoogleCloudPlatform/osconfig/util/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	agentendpointpb "github.com/GoogleCloudPlatform/osconfig/internal/google.golang.org/genproto/googleapis/cloud/osconfig/agentendpoint/v1alpha1"
)

var (
	aptInstalledPR = &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource{
		DesiredState: agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_INSTALLED,
		SystemPackage: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Apt{
			Apt: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_APT{Name: "foo"}}}
	aptRemovedPR = &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource{
		DesiredState: agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_REMOVED,
		SystemPackage: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Apt{
			Apt: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_APT{Name: "foo"}}}
	googetInstalledPR = &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource{
		DesiredState: agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_INSTALLED,
		SystemPackage: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Googet{
			Googet: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_GooGet{Name: "foo"}}}
	googetRemovedPR = &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource{
		DesiredState: agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_REMOVED,
		SystemPackage: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Googet{
			Googet: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_GooGet{Name: "foo"}}}
	yumInstalledPR = &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource{
		DesiredState: agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_INSTALLED,
		SystemPackage: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Yum{
			Yum: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_YUM{Name: "foo"}}}
	yumRemovedPR = &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource{
		DesiredState: agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_REMOVED,
		SystemPackage: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Yum{
			Yum: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_YUM{Name: "foo"}}}
	zypperInstalledPR = &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource{
		DesiredState: agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_INSTALLED,
		SystemPackage: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Zypper_{
			Zypper: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Zypper{Name: "foo"}}}
	zypperRemovedPR = &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource{
		DesiredState: agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_REMOVED,
		SystemPackage: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Zypper_{
			Zypper: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Zypper{Name: "foo"}}}
)

func TestPackageResourceValidate(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	tmpFile := filepath.Join(tmpDir, "foo")
	if err := ioutil.WriteFile(tmpFile, nil, 0644); err != nil {
		t.Fatal(err)
	}
	var tests = []struct {
		name    string
		wantErr bool
		prpb    *agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource
		wantMP  ManagedPackage
	}{
		{
			"Blank",
			true,
			&agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource{},
			ManagedPackage{},
		},
		{
			"AptInstalled",
			false,
			aptInstalledPR,
			ManagedPackage{Apt: &AptPackage{
				DesiredState:    agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_INSTALLED,
				PackageResource: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_APT{Name: "foo"}}},
		},
		{
			"AptRemoved",
			false,
			aptRemovedPR,
			ManagedPackage{Apt: &AptPackage{
				DesiredState:    agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_REMOVED,
				PackageResource: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_APT{Name: "foo"}}},
		},
		{
			"DebInstalled",
			false,
			&agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource{
				DesiredState: agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_INSTALLED,
				SystemPackage: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Deb_{
					Deb: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Deb{
						Source: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_File{
							File: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_File_LocalPath{LocalPath: tmpFile}}}}},
			ManagedPackage{Deb: &DebPackage{
				PackageResource: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Deb{
					Source: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_File{
						File: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_File_LocalPath{LocalPath: tmpFile}}}}},
		},
		{
			"DebRemoved",
			true,
			&agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource{
				DesiredState:  agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_REMOVED,
				SystemPackage: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Deb_{},
			},
			ManagedPackage{},
		},
		{
			"GoGetInstalled",
			false,
			googetInstalledPR,
			ManagedPackage{GooGet: &GooGetPackage{
				DesiredState:    agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_INSTALLED,
				PackageResource: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_GooGet{Name: "foo"}}},
		},
		{
			"GooGetRemoved",
			false,
			googetRemovedPR,
			ManagedPackage{GooGet: &GooGetPackage{
				DesiredState:    agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_REMOVED,
				PackageResource: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_GooGet{Name: "foo"}}},
		},
		{
			"MSIInstalled",
			false,
			&agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource{
				DesiredState: agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_INSTALLED,
				SystemPackage: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Msi{
					Msi: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_MSI{
						Source: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_File{
							File: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_File_LocalPath{LocalPath: tmpFile}}}}},
			ManagedPackage{MSI: &MSIPackage{
				PackageResource: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_MSI{
					Source: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_File{
						File: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_File_LocalPath{LocalPath: tmpFile}}}}},
		},
		{
			"MSIRemoved",
			true,
			&agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource{
				DesiredState:  agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_REMOVED,
				SystemPackage: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Msi{},
			},
			ManagedPackage{},
		},
		{
			"YumInstalled",
			false,
			yumInstalledPR,
			ManagedPackage{Yum: &YumPackage{
				DesiredState:    agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_INSTALLED,
				PackageResource: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_YUM{Name: "foo"}}},
		},
		{
			"YumRemoved",
			false,
			yumRemovedPR,
			ManagedPackage{Yum: &YumPackage{
				DesiredState:    agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_REMOVED,
				PackageResource: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_YUM{Name: "foo"}}},
		},
		{
			"ZypperInstalled",
			false,
			zypperInstalledPR,
			ManagedPackage{Zypper: &ZypperPackage{
				DesiredState:    agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_INSTALLED,
				PackageResource: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Zypper{Name: "foo"}}},
		},
		{
			"ZypperRemoved",
			false,
			zypperRemovedPR,
			ManagedPackage{Zypper: &ZypperPackage{
				DesiredState:    agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_REMOVED,
				PackageResource: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Zypper{Name: "foo"}}},
		},
		{
			"RPMInstalled",
			false,
			&agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource{
				DesiredState: agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_INSTALLED,
				SystemPackage: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Rpm{
					Rpm: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_RPM{
						Source: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_File{
							File: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_File_LocalPath{LocalPath: tmpFile}}}}},
			ManagedPackage{RPM: &RPMPackage{
				PackageResource: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_RPM{
					Source: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_File{
						File: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_File_LocalPath{LocalPath: tmpFile}}}}},
		},
		{
			"RPMRemoved",
			true,
			&agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource{
				DesiredState:  agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_REMOVED,
				SystemPackage: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource_Rpm{},
			},
			ManagedPackage{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &OSPolicyResource{
				ApplyConfigTask_OSPolicy_Resource: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource{
					ResourceType: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_Pkg{Pkg: tt.prpb},
				},
			}
			err := pr.Validate(ctx)
			if err != nil && !tt.wantErr {
				t.Fatalf("Unexpected error: %v", err)
			}
			if err == nil && tt.wantErr {
				t.Fatal("Expected error and did not get one.")
			}

			wantMR := &ManagedResources{Packages: []ManagedPackage{tt.wantMP}}
			if err != nil {
				wantMR = nil
			}

			opts := []cmp.Option{protocmp.Transform(), cmp.AllowUnexported(ManagedPackage{}), cmp.AllowUnexported(DebPackage{}), cmp.AllowUnexported(RPMPackage{}), cmp.AllowUnexported(MSIPackage{})}
			if diff := cmp.Diff(pr.ManagedResources(), wantMR, opts...); diff != "" {
				t.Errorf("OSPolicyResource does not match expectation: (-got +want)\n%s", diff)
			}
			if diff := cmp.Diff(pr.resource.(*packageResouce).managedPackage, tt.wantMP, opts...); diff != "" {
				t.Errorf("packageResouce does not match expectation: (-got +want)\n%s", diff)
			}
		})
	}
}

func TestPopulateInstalledCache(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockCommandRunner := utilmocks.NewMockCommandRunner(mockCtrl)
	packages.SetCommandRunner(mockCommandRunner)
	mockCommandRunner.EXPECT().Run(ctx, exec.Command("googet.exe", "installed")).Return([]byte("Installed Packages:\nfoo.x86_64 1.2.3@4\nbar.noarch 1.2.3@4"), nil, nil).Times(1)

	if err := populateInstalledCache(ctx, ManagedPackage{GooGet: &GooGetPackage{}}); err != nil {
		t.Fatalf("Unexpected error from populateInstalledCache: %v", err)
	}

	want := map[string]struct{}{"foo": {}, "bar": {}}
	if diff := cmp.Diff(gooInstalled.cache, want); diff != "" {
		t.Errorf("OSPolicyResource does not match expectation: (-got +want)\n%s", diff)
	}
}

func TestPackageResourceCheckState(t *testing.T) {
	ctx := context.Background()
	var tests = []struct {
		name               string
		installedCache     map[string]struct{}
		cachePointer       *packageCache
		prpb               *agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource
		wantInDesiredState bool
	}{
		// We only need to test the full set once as all the logic is shared.
		{
			"AptInstalledNeedsInstalled",
			map[string]struct{}{"foo": {}},
			&aptInstalled,
			aptInstalledPR,
			true,
		},
		{
			"AptInstalledNeedsRemoved",
			map[string]struct{}{"foo": {}},
			&aptInstalled,
			aptRemovedPR,
			false,
		},
		{
			"AptRemovedNeedsInstalled",
			map[string]struct{}{},
			&aptInstalled,
			aptInstalledPR,
			false,
		},
		{
			"AptRemovedNeedsRemoved",
			map[string]struct{}{},
			&aptInstalled,
			aptRemovedPR,
			true,
		},

		// For the rest of the package types we only need to test one scenario.
		{
			"GooGetInstalledNeedsInstalled",
			map[string]struct{}{"foo": {}},
			&gooInstalled,
			googetInstalledPR,
			true,
		},
		{
			"YUMInstalledNeedsInstalled",
			map[string]struct{}{"foo": {}},
			&yumInstalled,
			yumInstalledPR,
			true,
		},
		{
			"ZypperInstalledNeedsInstalled",
			map[string]struct{}{"foo": {}},
			&zypperInstalled,
			zypperInstalledPR,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &OSPolicyResource{
				ApplyConfigTask_OSPolicy_Resource: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource{
					ResourceType: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_Pkg{Pkg: tt.prpb},
				},
			}
			// Run validate first to make sure everything gets setup correctly.
			// This adds complexity to this 'unit' test and turns it into more
			// of a integration test but reduces overall test functions and gives
			// us good coverage.
			if err := pr.Validate(ctx); err != nil {
				t.Fatalf("Unexpected Validate error: %v", err)
			}

			tt.cachePointer.cache = tt.installedCache
			tt.cachePointer.refreshed = time.Now()
			if err := pr.CheckState(ctx); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.wantInDesiredState != pr.InDesiredState() {
				t.Fatalf("Unexpected InDesiredState, want: %t, got: %t", tt.wantInDesiredState, pr.InDesiredState())
			}
		})
	}
}

func TestPackageResourceEnforceState(t *testing.T) {
	ctx := context.Background()
	var tests = []struct {
		name        string
		prpb        *agentendpointpb.ApplyConfigTask_OSPolicy_Resource_PackageResource
		expectedCmd *exec.Cmd
	}{
		{
			"AptInstalled",
			aptInstalledPR,
			func() *exec.Cmd {
				cmd := exec.Command("/usr/bin/apt-get", "install", "-y", "foo")
				cmd.Env = append(os.Environ(),
					"DEBIAN_FRONTEND=noninteractive",
				)
				return cmd
			}(),
		},
		{
			"AptRemoved",
			aptRemovedPR,
			func() *exec.Cmd {
				cmd := exec.Command("/usr/bin/apt-get", "remove", "-y", "foo")
				cmd.Env = append(os.Environ(),
					"DEBIAN_FRONTEND=noninteractive",
				)
				return cmd
			}(),
		},
		{
			"GooGetInstalled",
			googetInstalledPR,
			exec.Command("googet.exe", "-noconfirm", "install", "foo"),
		},
		{
			"GooGetRemoved",
			googetRemovedPR,
			exec.Command("googet.exe", "-noconfirm", "remove", "foo"),
		},
		{
			"YumInstalled",
			yumInstalledPR,
			exec.Command("/usr/bin/yum", "install", "--assumeyes", "foo"),
		},
		{
			"YumRemoved",
			yumRemovedPR,
			exec.Command("/usr/bin/yum", "remove", "--assumeyes", "foo"),
		},
		{
			"ZypperInstalled",
			zypperInstalledPR,
			exec.Command("/usr/bin/zypper", "--gpg-auto-import-keys", "--non-interactive", "install", "--auto-agree-with-licenses", "foo"),
		},
		{
			"ZypperRemoved",
			zypperRemovedPR,
			exec.Command("/usr/bin/zypper", "--non-interactive", "remove", "foo"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &OSPolicyResource{
				ApplyConfigTask_OSPolicy_Resource: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource{
					ResourceType: &agentendpointpb.ApplyConfigTask_OSPolicy_Resource_Pkg{Pkg: tt.prpb},
				},
			}
			// Run Validate first to make sure everything gets setup correctly.
			// This adds complexity to this 'unit' test and turns it into more
			// of a integration test but reduces overall test functions and gives
			// us good coverage.
			if err := pr.Validate(ctx); err != nil {
				t.Fatalf("Unexpected Validate error: %v", err)
			}

			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockCommandRunner := utilmocks.NewMockCommandRunner(mockCtrl)
			packages.SetCommandRunner(mockCommandRunner)
			mockCommandRunner.EXPECT().Run(ctx, tt.expectedCmd)

			if err := pr.EnforceState(ctx); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}
