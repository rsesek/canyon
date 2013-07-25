// Copyright (c) 2013 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

/*
canyon is a tool for making big splits. This assumes that you have a very large
changelist prepared in a single branch. canyon will then split up the large
change into multiple branches, by OWNERS file, and then will prepare a changelist
description.
*/

const kMaxDepth = 1

// A changeList represents one of the split changelists.
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

func main() {
	branch := strings.TrimSpace(gitOrDie("symbolic-ref", "--short", "HEAD"))

	fmt.Printf("Split changelist on branch %q into sub-changelists? [y/N] ", branch)
	buf := make([]byte, 1)
	os.Stdin.Read(buf)
	if buf[0] != 'y' {
		fmt.Println("Exiting")
		return
	}

	log.Print("Gathering changed files")
	files := strings.Split(gitOrDie("diff", "--name-only", "origin/master"), "\n")

	log.Print("Splitting changed files into groups for changelists")
	splits := make(map[string]*changeList)
	for _, file := range files {
		if file == "" {
			continue
		}
		splitForFile(splits, file)
	}

	log.Print("Creating branches for splits")
	for _, cl := range splits {
		splitBranch := cl.branchName(branch)
		log.Printf("Preparing branch %s", splitBranch)

		_, err := git("checkout", "-b", splitBranch, "origin/master")
		if err != nil {
			log.Printf("Failed to create new branch %q: %v", splitBranch, err)
			continue
		}

		_, err = git("checkout", branch, cl.base)
		if err != nil {
			log.Print("Failed to check out subdirectory from root branch")
			gitOrDie("reset", "--hard", "origin/master")
			continue
		}

		desc := fmt.Sprintf("Update include paths in %s for base/process changes.\n\nBUG=242290", cl.base)
		desc += "\n\n===== Affected paths: =====\n"
		for _, file := range cl.paths {
			desc += fmt.Sprintf("%s\n", file)
		}
		desc += "\n" + cl.description

		_, err = git("commit", "-a", "-m", desc)
		if err != nil {
			log.Print("Failed to create subchangelist")
			gitOrDie("reset", "--hard", "origin/master")
			continue
		}
	}

	git("checkout", branch)
}

// git runs the specified git commands and returns the output as a string,
// blocking to completion.
func git(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(stdout), nil
}

// gitOrDie runs the git command and panics on failure.
func gitOrDie(args ...string) string {
	r, e := git(args...)
	if e != nil {
		panic(e.Error())
	}
	return r
}

// splitForFile returns a changeList object for the given file path.
func splitForFile(splits map[string]*changeList, file string) *changeList {
	base := path.Dir(file)
	for {
		if cl, ok := splits[base]; ok {
			cl.addPath(file)
			return cl
		}

		owners := path.Join(base, "OWNERS")
		f, err := os.Open(owners)
		if err != nil || strings.Count(base, string(os.PathSeparator)) > kMaxDepth {
			base = path.Dir(base)
			continue
		} else {
			cl := &changeList{
				base: base,
				description: fmt.Sprintf("===== Contents of %s =====\n", owners),
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
			splits[base] = cl
			return cl
		}
	}
}
