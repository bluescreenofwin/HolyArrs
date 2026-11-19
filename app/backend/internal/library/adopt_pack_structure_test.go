package library

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	domain "github.com/HolyArrs/holyarrs/app/backend/internal/core/model"
	"github.com/HolyArrs/holyarrs/app/backend/internal/metadata"
)

func TestPackStructuralDiscovery(t *testing.T) {
	root := t.TempDir()
	outer := "The Expendables 1, 2, 3 - Sylvester Stallone Action 2010-2014 Eng Sub 720p [H264-mp4]"
	inner := outer + "/The Expendables 1, 2, 3 [H264-mp4]"
	files := []string{
		inner + "/01 The Expendables - Extended Directors Cut Sylvester Stallone 2010 Eng Sub 720p [H264-mp4].mp4",
		inner + "/02 The Expendables 2 - Sylvester Stallone 2012 Eng Sub 720p [H264-mp4].mp4",
		inner + "/03 The Expendables 3 - Extended Unrated Sylvester Stallone 2014 Eng Sub 720p [H264-mp4].mp4",
	}
	for _, f := range files {
		writeFile(t, filepath.Join(root, f), "original video")
	}
	writeFile(t, filepath.Join(root, outer, "Images", "cover.jpg"), "cover")
	writeFile(t, filepath.Join(root, outer, "Subtitles Eng [SubRip - MicroDVD]", "title.srt"), "subtitle")
	es, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	got := movieCandidates(root, es, nil)
	if len(got) != 3 {
		t.Fatalf("want 3 files, got %q", got)
	}
	queries := []string{"The Expendables Sylvester Stallone", "The Expendables 2 Sylvester Stallone", "The Expendables 3 Sylvester Stallone"}
	for i, f := range files {
		if got[i] != f {
			t.Fatalf("changed candidate: %q", got)
		}
		name, year := FolderName(candidateLookup(f))
		if name != queries[i] || year != 2010+i*2 {
			t.Fatalf("lookup %q/%d", name, year)
		}
		b, e := os.ReadFile(filepath.Join(root, f))
		if e != nil || string(b) != "original video" {
			t.Fatal("file changed")
		}
	}
}

func TestGenericNestedPackStructure(t *testing.T) {
	root := t.TempDir()
	for _, f := range []string{"Example 2000-2002/Films/01 Example - Extended Directors Cut Pat Smith 2000 Eng 720p.mkv", "Example 2000-2002/Films/02 Example 2 - Unrated Pat Smith 2002 Eng 720p.mkv"} {
		writeFile(t, filepath.Join(root, f), "video")
	}
	es, _ := os.ReadDir(root)
	got := movieCandidates(root, es, nil)
	if len(got) != 2 {
		t.Fatalf("%q", got)
	}
	name, year := FolderName(candidateLookup(got[0]))
	if name != "Example Pat Smith" || year != 2000 {
		t.Fatalf("%s %d", name, year)
	}
}

func TestAmbiguousPackCannotAdoptLargestVideo(t *testing.T) {
	s := open(t)
	lf := addFolder(t, s)
	for _, f := range []string{"Unknown Collection/one.mkv", "Unknown Collection/two.mkv"} {
		writeFile(t, filepath.Join(lf.Path, f), "video")
	}
	svc := &Service{Store: s, Metadata: adoptTMDB}
	p, err := PlanAdopt(context.Background(), s.Read(), adoptTMDB, lf.ID)
	if err != nil || len(p.Folders) != 1 || p.Folders[0].Main != "" || p.Folders[0].Match != nil || p.Folders[0].Why != "Possible collection pack: review its contents and import individual movie files." {
		t.Fatalf("%+v %v", p, err)
	}
	r, err := svc.Adopt(context.Background(), lf.ID, []AdoptChoice{{Folder: "Unknown Collection", TMDB: "949"}}, domain.System)
	if err != nil || len(r) != 1 || r[0].Error == "" || r[0].TitleID != 0 {
		t.Fatalf("adopted ambiguous pack: %+v %v", r, err)
	}
}

func TestPackGuardPreservesEncodesAndDiscs(t *testing.T) {
	for _, videos := range [][]AdoptVideo{
		{{Path: "Example Collection/Example (2000) 1080p.mkv"}, {Path: "Example Collection/Example (2000) 2160p.mkv"}},
		{{Path: "Example Collection/Disc 1/movie.mkv"}, {Path: "Example Collection/Disc 2/movie.mkv"}},
	} {
		if possibleCollectionPack("Example Collection", videos) {
			t.Fatalf("ordinary movie split: %+v", videos)
		}
	}
}

func TestPackPersonMustBeVerified(t *testing.T) {
	m := evidenceMovie("Example Film", 2000)
	detail := m
	detail.People = []string{"Pat Smith"}
	p := &evidenceProvider{results: map[string][]metadata.Movie{"Example Film Pat Smith/2000": {m}}, detail: detail}
	_, _, ms, err := matchCandidate(context.Background(), p, "Example Collection/01 Example Film - Extended Unrated Pat Smith 2000 Eng 720p.mkv")
	if err != nil || len(ms) != 1 || ms[0].Score != 1 || p.details != 1 {
		t.Fatalf("%+v %v", ms, err)
	}
}
