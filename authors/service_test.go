package authors

import (
	"bufio"
	"encoding/json"
	"github.com/Financial-Times/tme-reader/tmereader"
	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"
)

type testSuiteForAuthors struct {
	name  string
	uuid  string
	found bool
	err   error
}

func TestInit(t *testing.T) {
	repo := blockingRepo{}
	repo.Add(1)
	tmpfile := getTempFile(t)
	defer os.Remove(tmpfile.Name())
	service := createTestAuthorService(&repo, tmpfile.Name())
	defer func() {
		repo.Done()
		service.Shutdown()
	}()
	assert.False(t, service.isDataLoaded())
	assert.True(t, service.isInitialised())
}

func TestGetAuthors(t *testing.T) {
	tmpfile := getTempFile(t)
	defer os.Remove(tmpfile.Name())
	repo := dummyRepo{terms: []term{{CanonicalName: "Bob", RawID: "bob"}, {CanonicalName: "Fred", RawID: "fred"}}}
	service := createTestAuthorService(&repo, tmpfile.Name())
	defer service.Shutdown()
	waitTillInit(t, service)
	waitTillDataLoaded(t, service)
	pv, err := service.getAuthors()

	var wg sync.WaitGroup
	var res []author
	wg.Add(1)
	go func(reader io.Reader, w *sync.WaitGroup) {
		var err error
		scan := bufio.NewScanner(reader)
		for scan.Scan() {
			var p author
			assert.NoError(t, err)
			err = json.Unmarshal(scan.Bytes(), &p)
			assert.NoError(t, err)
			res = append(res, p)
		}
		wg.Done()
	}(&pv, &wg)
	wg.Wait()

	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "28d66fcc-bb56-363d-80c1-f2d957ef58cf", res[0].UUID)
	assert.Equal(t, "be2e7e2b-0fa2-3969-a69b-74c46e754032", res[1].UUID)
}

func TestGetAuthorUUIDs(t *testing.T) {
	tmpfile := getTempFile(t)
	defer os.Remove(tmpfile.Name())
	repo := dummyRepo{terms: []term{{CanonicalName: "Bob", RawID: "bob"}, {CanonicalName: "Fred", RawID: "fred"}}}
	service := createTestAuthorService(&repo, tmpfile.Name())
	defer service.Shutdown()
	waitTillInit(t, service)
	waitTillDataLoaded(t, service)
	pv, err := service.getAuthorUUIDs()

	var wg sync.WaitGroup
	var res []authorUUID
	wg.Add(1)
	go func(reader io.Reader, w *sync.WaitGroup) {
		var err error
		scan := bufio.NewScanner(reader)
		for scan.Scan() {
			var p authorUUID
			assert.NoError(t, err)
			err = json.Unmarshal(scan.Bytes(), &p)
			assert.NoError(t, err)
			res = append(res, p)
		}
		wg.Done()
	}(&pv, &wg)
	wg.Wait()

	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "28d66fcc-bb56-363d-80c1-f2d957ef58cf", res[0].UUID)
	assert.Equal(t, "be2e7e2b-0fa2-3969-a69b-74c46e754032", res[1].UUID)
}

func TestGetAuthorLinks(t *testing.T) {
	tmpfile := getTempFile(t)
	defer os.Remove(tmpfile.Name())
	repo := dummyRepo{terms: []term{{CanonicalName: "Bob", RawID: "bob"}, {CanonicalName: "Fred", RawID: "fred"}}}
	service := createTestAuthorService(&repo, tmpfile.Name())
	defer service.Shutdown()
	waitTillInit(t, service)
	waitTillDataLoaded(t, service)
	pv, err := service.getAuthorLinks()

	var wg sync.WaitGroup
	var res []authorLink
	wg.Add(1)
	go func(reader io.Reader, w *sync.WaitGroup) {
		var err error
		jsonBlob, err := ioutil.ReadAll(reader)
		assert.NoError(t, err)
		log.Infof("Got bytes: %v", string(jsonBlob[:]))
		err = json.Unmarshal(jsonBlob, &res)
		assert.NoError(t, err)
		wg.Done()
	}(&pv, &wg)
	wg.Wait()

	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "/base/url/28d66fcc-bb56-363d-80c1-f2d957ef58cf", res[0].APIURL)
	assert.Equal(t, "/base/url/be2e7e2b-0fa2-3969-a69b-74c46e754032", res[1].APIURL)
}

func TestGetCount(t *testing.T) {
	tmpfile := getTempFile(t)
	defer os.Remove(tmpfile.Name())
	repo := dummyRepo{terms: []term{{CanonicalName: "Bob", RawID: "bob"}, {CanonicalName: "Fred", RawID: "fred"}}}
	service := createTestAuthorService(&repo, tmpfile.Name())
	defer service.Shutdown()
	waitTillInit(t, service)
	waitTillDataLoaded(t, service)
	assertCount(t, service, 2)
}

func TestReload(t *testing.T) {
	tmpfile := getTempFile(t)
	defer os.Remove(tmpfile.Name())
	repo := dummyRepo{terms: []term{{CanonicalName: "Bob", RawID: "bob"}, {CanonicalName: "Fred", RawID: "fred"}}}
	service := createTestAuthorService(&repo, tmpfile.Name())
	defer service.Shutdown()
	waitTillInit(t, service)
	waitTillDataLoaded(t, service)
	assertCount(t, service, 2)
	repo.terms = append(repo.terms, term{CanonicalName: "Third", RawID: "third"})
	repo.count = 0
	assert.NoError(t, service.reloadDB())
	waitTillInit(t, service)
	waitTillDataLoaded(t, service)
	assertCount(t, service, 3)
}

func TestGetAuthorByUUID(t *testing.T) {
	tmpfile := getTempFile(t)
	defer os.Remove(tmpfile.Name())
	repo := dummyRepo{terms: []term{{CanonicalName: "Bob", RawID: "bob"}, {CanonicalName: "Fred", RawID: "fred"}}}
	service := createTestAuthorService(&repo, tmpfile.Name())
	defer service.Shutdown()
	waitTillInit(t, service)
	waitTillDataLoaded(t, service)

	tests := []testSuiteForAuthors{
		{"Success", "28d66fcc-bb56-363d-80c1-f2d957ef58cf", true, nil},
		{"Success", "xxxxxxxx-bb56-363d-80c1-f2d957ef58cf", false, nil}}
	for _, test := range tests {
		author, found, err := service.getAuthorByUUID(test.uuid)
		if test.err != nil {
			assert.Equal(t, test.err, err)
		} else if test.found {
			assert.True(t, found)
			assert.NotNil(t, author)
		} else {
			assert.False(t, found)
		}
	}
}

func TestFailingOpeningDB(t *testing.T) {
	dir, err := ioutil.TempDir("", "service_test")
	assert.NoError(t, err)
	service := createTestAuthorService(&dummyRepo{}, dir)
	defer service.Shutdown()
	for i := 1; i <= 1000; i++ {
		if !service.isInitialised() {
			log.Info("isInitialised was false")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.False(t, service.isInitialised(), "isInitialised should be false")
}

func assertCount(t *testing.T, s AuthorService, expected int) {
	count, err := s.getCount()
	assert.NoError(t, err)
	assert.Equal(t, expected, count)
}

func createTestAuthorService(repo tmereader.Repository, cacheFileName string) AuthorService {
	return NewAuthorService(repo, "/base/url", "taxonomy_string", 1, cacheFileName)
}

func getTempFile(t *testing.T) *os.File {
	tmpfile, err := ioutil.TempFile("", "example")
	assert.NoError(t, err)
	assert.NoError(t, tmpfile.Close())
	log.Debug("File:%s", tmpfile.Name())
	return tmpfile
}

func waitTillInit(t *testing.T, s AuthorService) {
	for i := 1; i <= 1000; i++ {
		if s.isInitialised() {
			log.Info("isInitialised was true")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.True(t, s.isInitialised())
}

func waitTillDataLoaded(t *testing.T, s AuthorService) {
	for i := 1; i <= 1000; i++ {
		if s.isDataLoaded() {
			log.Info("isDataLoaded was true")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.True(t, s.isDataLoaded())
}

type dummyRepo struct {
	sync.Mutex
	terms []term
	err   error
	count int
}

func (d *dummyRepo) GetTmeTermsFromIndex(startRecord int) ([]interface{}, error) {
	defer func() {
		d.count++
	}()
	if len(d.terms) == d.count {
		return nil, d.err
	}
	return []interface{}{d.terms[d.count]}, d.err
}

// Never used
func (d *dummyRepo) GetTmeTermById(uuid string) (interface{}, error) {
	return nil, nil
}

type blockingRepo struct {
	sync.WaitGroup
	err  error
	done bool
}

func (d *blockingRepo) GetTmeTermsFromIndex(startRecord int) ([]interface{}, error) {
	d.Wait()
	if d.done {
		return nil, d.err
	}
	d.done = true
	return []interface{}{term{CanonicalName: "Bob", RawID: "bob"}}, d.err
}

// Never used
func (d *blockingRepo) GetTmeTermById(uuid string) (interface{}, error) {
	return nil, nil
}
