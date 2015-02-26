package cli

import (
	"fmt"
	"net/url"

	"github.com/tyba/opensource-search/sources/social/http"
	"github.com/tyba/opensource-search/sources/social/readers"

	"gopkg.in/mgo.v2"
)

type Bitbucket struct {
	MongoDBHost string `short:"m" long:"mongo" default:"localhost" description:"mongodb hostname"`
	MaxThreads  int    `short:"t" long:"threads" default:"4" description:"number of t"`

	bitbucket *readers.BitbucketReader
	storage   *mgo.Collection
}

func (b *Bitbucket) Execute(args []string) error {
	session, _ := mgo.Dial("mongodb://" + b.MongoDBHost)

	b.bitbucket = readers.NewBitbucketReader(http.NewClient(true))
	b.storage = session.DB("bitbucket").C("repositories")

	r, err := b.bitbucket.GetRepositories(url.Values{})
	if err != nil {
		return err
	}

	for {
		r, err = b.bitbucket.GetRepositories(r.Next.Query())
		if err != nil {
			return err
		}

		b.saveBitbucketPagedResult(r)
	}

	return nil
}

func (b *Bitbucket) saveBitbucketPagedResult(res *readers.BitbucketPagedResult) error {
	fmt.Printf("Retrieved: %d repositorie(s)\nNext: %s\n", len(res.Values), res.Next)

	for _, r := range res.Values {
		if err := b.storage.Insert(r); err != nil {
			return err
		}
	}

	return nil
}
