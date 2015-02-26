package cli

import (
	"fmt"

	"github.com/tyba/opensource-search/sources/social/http"
	"github.com/tyba/opensource-search/sources/social/readers"
	"github.com/tyba/opensource-search/types/social"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type LinkedIn struct {
	MongoDBHost string `short:"m" long:"mongo" default:"localhost" description:"mongodb hostname"`

	linkedin *readers.LinkedInReader
	augur    *mgo.Collection
	storage  *mgo.Collection
}

type augurData struct {
	Profiles struct {
		LinkedInURL string `bson:"linkedin_url"`
		GithubURL   string `bson:"github_url"`
		TwitterURL  string `bson:"twitter_url"`
	}
}

func (l *LinkedIn) Execute(args []string) error {
	session, _ := mgo.Dial("mongodb://" + l.MongoDBHost)

	l.linkedin = readers.NewLinkedInReader(http.NewCachedClient(session))
	l.storage = session.DB("social").C("linkedin")
	l.augur = session.DB("social").C("augur")

	pending := l.get()
	for {
		result := &augurData{}
		if !pending.Next(result) {
			break
		}

		l.processData(result)
	}

	return nil
}

func (l *LinkedIn) get() *mgo.Iter {
	q := bson.M{
		"profiles.linkedin_url": bson.M{
			"$exists": 1,
		},
		"crawler.linkedin_url": bson.M{
			"$exists": 0,
		},
	}

	return l.augur.Find(q).Sort("-_id").Iter()
}

func (l *LinkedIn) processData(d *augurData) {
	url := d.Profiles.LinkedInURL
	if l.has(url) {
		fmt.Printf("SKIP: %q\n", url)
		l.done(url, 200)

		return
	}

	p, err := l.linkedin.GetProfileByURL(url)
	if err != nil {
		fmt.Printf("ERROR: %q, %s\n", url, err)
		l.done(url, 500)

		return
	}

	l.saveLinkedInProfile(p)
	fmt.Printf("DONE: %s\n", p.FullName)
	l.done(url, 200)

	return
}

func (l *LinkedIn) has(url string) bool {
	q := bson.M{"url": url}

	if c, _ := l.storage.Find(q).Count(); c == 0 {
		return false
	}

	return true
}

func (l *LinkedIn) done(url string, status int) {
	q := bson.M{"profiles.linkedin_url": url}
	s := bson.M{
		"$set": bson.M{
			"crawler.linkedin_url": 200,
		},
	}

	_, err := l.augur.UpdateAll(q, s)
	if err != nil {
		panic(err)
	}
}

func (l *LinkedIn) saveLinkedInProfile(p *social.LinkedInProfile) error {
	return l.storage.Insert(p)
}
