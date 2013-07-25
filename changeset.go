// Copyright (c) 2013 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

// A changeSet is the result of splitting one branch into several.
type changeSet struct {
	// The original branch name.
	branch string

	// A map of directory base paths to changeList objects, which represent
	// how |branch| is split.
	splits map[string]*changeList
}

func newChangeSet(branch string) *changeSet {
	return &changeSet{
		branch: branch,
		splits: make(map[string]*changeList),
	}
}

// splitByFile splits |file| in the changeSet into a changeList by looking for
// a shared |splitBy| file in a parent directory.
func (cs *changeSet) splitByFile(splitBy string, file string) *changeList {
	base := path.Dir(file)
	for {
		if cl, ok := cs.splits[base]; ok {
			cl.addPath(file)
			return cl
		}

		splitByPath := path.Join(base, splitBy)
		f, err := os.Open(splitByPath)
		if err != nil || strings.Count(base, string(os.PathSeparator)) > kMaxDepth {
			base = path.Dir(base)
			continue
		} else {
			cl := &changeList{
				base:        base,
				description: fmt.Sprintf("===== Contents of %s =====\n", splitByPath),
			}
			bio := bufio.NewReader(f)
			for {
				line, err := bio.ReadString('\n')
				if err != nil {
					break
				}
				cl.description += line
			}
			f.Close()

			cl.addPath(file)
			cs.splits[base] = cl
			return cl
		}
	}
}

// A changeList represents one of the split branches.
type changeList struct {
	// The base path for this changelist.
	base string

	// List of modified paths.
	paths []string

	// Extra data for the CL description.
	description string
}

func (cl *changeList) addPath(p string) {
	cl.paths = append(cl.paths, p)
}

func (cl *changeList) branchName(root string) string {
	normalize := strings.Replace(cl.base, "/", "-", -1)
	return fmt.Sprintf("canyon/%s/%s", root, normalize)
}

func (cl *changeList) String() string {
	return fmt.Sprintf("<changeList %p : %d files>", cl, len(cl.paths))
}
