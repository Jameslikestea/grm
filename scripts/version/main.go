package main

import (
	"log"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/leodido/go-conventionalcommits"
	"github.com/leodido/go-conventionalcommits/parser"
)

type Tag struct {
	version *semver.Version
	hash    plumbing.Hash
}

func main() {
	r := getRepo()
	h := getHead(r)
	t := getTags(r)

	seen := map[plumbing.Hash]bool{}
	messages := []string{}

	log.Printf("HEAD @ %s", h.String())
	for _, tag := range t {
		markSeen(r, tag.hash, seen)
	}

	getNewCommits(r, h, seen, &messages)

	m := parser.NewMachine(parser.WithTypes(conventionalcommits.TypesConventional))

	for _, msg := range messages {
		res, err := m.Parse([]byte(strings.TrimSpace(msg)))
		if err != nil {
			continue
		}
		p := res.(*conventionalcommits.ConventionalCommit)
		log.Println(p.Type)
	}
}

func getRepo() *git.Repository {
	r, err := git.PlainOpen(".git")
	if err != nil {
		log.Fatalln(err)
	}

	return r
}

func getHead(r *git.Repository) plumbing.Hash {
	currentCommit, err := r.Head()
	if err != nil {
		log.Fatalln(err)
	}

	return currentCommit.Hash()
}

func getTags(r *git.Repository) []Tag {
	iter, err := r.Tags()
	if err != nil {
		log.Fatalln(err)
	}
	var pr []Tag
	iter.ForEach(
		func(r *plumbing.Reference) error {
			if v := checkValidVersion(r.Name().Short()); v != "" {
				ver := semver.New(v)
				pr = append(
					pr, Tag{
						version: ver,
						hash:    r.Hash(),
					},
				)
			}
			return nil
		},
	)

	return pr
}

func checkValidVersion(name string) string {
	if !strings.HasPrefix(name, "v") {
		return ""
	}

	name = name[1:]
	r := regexp.MustCompile("^([0-9]+)\\.([0-9]+)\\.([0-9]+)(?:-([0-9A-Za-z-]+(?:\\.[0-9A-Za-z-]+)*))?(?:\\+[0-9A-Za-z-]+)?$")
	if r.Match([]byte(name)) {
		return name
	}
	return ""
}

func markSeen(r *git.Repository, h plumbing.Hash, seen map[plumbing.Hash]bool) {
	seen[h] = true
	c, err := r.CommitObject(h)
	if err != nil {
		log.Fatalln(err)
	}

	for _, hash := range c.ParentHashes {
		markSeen(r, hash, seen)
	}
}

func getNewCommits(r *git.Repository, h plumbing.Hash, seen map[plumbing.Hash]bool, messages *[]string) {
	c, err := r.CommitObject(h)
	if err != nil {
		log.Fatalln(err)
	}

	if _, ok := seen[c.Hash]; ok {
		return
	}

	*messages = append(*messages, c.Message)

	for _, hash := range c.ParentHashes {
		getNewCommits(r, hash, seen, messages)
	}
}
