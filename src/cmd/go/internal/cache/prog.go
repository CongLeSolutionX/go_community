package cache

import (
	"bufio"
	"bytes"
	"cmd/go/internal/base"
	"cmd/go/internal/mmap"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"
)

// ProgCache implements Cache via JSON messages over stdin/stdout to a child
// helper process which can then implement whatever caching policy/mechanism it
// wants.
//
// See https://github.com/golang/go/issues/59719
type ProgCache struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser  // from the child process
	stdin  io.WriteCloser // to the child process
	bw     *bufio.Writer  // to stdin
	jenc   *json.Encoder  // to bw

	// can are the commands that the child process declared that it supports.
	// This is effectively the versioning mechanism.
	can map[ProgCmd]bool

	// fuzzDirCache is another Cache implementation to use for the FuzzDir
	// method. In practice this is the default GOCACHE disk-based
	// implementation.
	//
	// TODO(bradfitz): maybe this isn't ideal. But we'd need to extend the Cache
	// interface and the fuzzing callers to be less disk-y to do more here.
	fuzzDirCache Cache

	mu       sync.Mutex // guards writing to the child process and following fields
	nextID   int64
	inFlight map[int64]chan<- *ProgResponse
}

// ProgCmd is a command that can be issued to a child process.
//
// If the interface needs to grow, we can add new commands or new versioned
// commands like "get2".
type ProgCmd string

const (
	cmdGet        = ProgCmd("get")
	cmdPut        = ProgCmd("put")
	cmdOutputFile = ProgCmd("output-file")
	cmdTrim       = ProgCmd("trim")
)

// ProgRequest is the JSON-encoded message that's sent from cmd/go to
// the GOCACHEPROG child process over stdin. Each JSON object is on its
// own line. A ProgRequest of Type "put" with BodySize > 0 will be followed
// by a line containing a base64-encoded JSON string literal of the body.
type ProgRequest struct {
	// ID is a unique number per process across all requests.
	// It must be echoed in the ProgResponse from the child.
	ID int64

	// Command is the type of request.
	// The cmd/go tool will only send commands that were declared
	// as supported by the child.
	Command ProgCmd

	// ActionID is non-nil for get and puts.
	ActionID []byte `json:",omitempty"` // or nil if not used

	// ObjectID is set for Type "put" and "output-file".
	ObjectID []byte `json:",omitempty"` // or nil if not used

	// WantDiskPath is whether the caller wants ProgResponse.DiskPath populated
	// for "get".
	WantDiskPath bool `json:",omitempty"`

	// Body is the body for "put" requests. It's sent after the JSON object
	// as a base64-encoded JSON string when BodySize is non-zero.
	// It's sent as a separate JSON value instead of being a struct field
	// send in this JSON object so large values can be streamed in both directions.
	// The base64 string body of a ProgRequest will always be written
	// immediately after the JSON object and a newline.
	Body io.Reader `json:"-"`

	// BodySize is the number of bytes of Body. If zero, the body isn't written.
	BodySize int64 `json:",omitempty"`
}

// ProgResponse is the JSON response from the child process to cmd/go.
//
// With the exception of the first protocol message that the child writes to its
// stdout with ID==0 and KnownCommands populated, these are only sent in
// response to a ProgRequest from cmd/go.
//
// ProgResponses can be sent in any order. The ID must match the request they're
// replying to.
type ProgResponse struct {
	ID  int64  // that corresponds to ProgRequest; they can be answered out of order
	Err string `json:",omitempty"` // if non-empty, the error

	// KnownCommands is included in the first message that cache helper program
	// writes to stdout on startup (with ID==0). It includes the
	// ProgRequest.Command types that are supported by the program.
	//
	// This lets us extend the gracefully over time (adding "get2", etc), or
	// fail gracefully when needed. It also lets us verify the program
	// wants to be a cache helper.
	KnownCommands []ProgCmd `json:",omitempty"`

	// For Get requests
	Miss      bool   `json:",omitempty"` // cache miss
	OutputID  []byte `json:",omitempty"`
	Size      int64  `json:",omitempty"`
	TimeNanos int64  `json:",omitempty"`

	// DiskPath is the absolute path on disk of the object ID
	// referenced by a "get" request with WantDiskPath or the ProgRequest.ObjectID
	// of a "output-file" request.
	DiskPath string `json:",omitempty"`
}

// startCacheProg starts the prog binary and returns a Cache implementation that
// talks to it.
//
// It blocks a few seconds to wait for the child process to successfully start
// and advertise its capabilities.
func startCacheProg(prog string, fuzzDirCache Cache) Cache {
	if fuzzDirCache == nil {
		panic(0)
	}
	fi, err := os.Stat(prog)
	if err != nil {
		base.Fatalf("invalid GOCACHEPROG: %v", err)
	}
	if !fi.Mode().IsRegular() {
		base.Fatalf("GOCACHEPROG value %q is not a binary", prog)
	}
	cmd := exec.Command(prog)
	out, err := cmd.StdoutPipe()
	if err != nil {
		base.Fatalf("StdoutPipe to GOCACHEPROG: %v", err)
	}
	in, err := cmd.StdinPipe()
	if err != nil {
		base.Fatalf("StdinPipe to GOCACHEPROG: %v", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		base.Fatalf("error starting GOCACHEPROG program %q: %v", prog, err)
	}

	pc := &ProgCache{
		fuzzDirCache: fuzzDirCache,
		cmd:          cmd,
		stdout:       out,
		stdin:        in,
		bw:           bufio.NewWriter(in),
		inFlight:     make(map[int64]chan<- *ProgResponse),
	}

	// Register our interest in the initial protocol message from the child to
	// us, saying what it can do.
	capResc := make(chan *ProgResponse, 1)
	pc.inFlight[0] = capResc

	pc.jenc = json.NewEncoder(pc.bw)
	go pc.readLoop()

	// Give the child process a few seconds to report its capabilities. This
	// should be instant and not require any slow work by the program.
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case <-timer.C:
		base.Fatalf("timeout waiting for initial capabilities JSON from GOCACHEPROG %v", prog)
	case capRes := <-capResc:
		can := map[ProgCmd]bool{}
		for _, cmd := range capRes.KnownCommands {
			can[cmd] = true
		}
		if len(can) == 0 {
			base.Fatalf("GOCACHEPROG %v declared no supported commands", prog)
		}
		pc.can = can
	}

	return pc
}

func (c *ProgCache) readLoop() {
	jd := json.NewDecoder(c.stdout)
	for {
		res := new(ProgResponse)
		if err := jd.Decode(res); err != nil {
			base.Fatalf("error reading JSON from GOCACHEPROG: %v", err)
		}
		c.mu.Lock()
		ch, ok := c.inFlight[res.ID]
		delete(c.inFlight, res.ID)
		c.mu.Unlock()
		if ok {
			ch <- res
		}
	}
}

func (c *ProgCache) send(ctx context.Context, req *ProgRequest) (*ProgResponse, error) {
	resc := make(chan *ProgResponse, 1)
	if err := c.writeToChild(req, resc); err != nil {
		return nil, err
	}
	select {
	case res := <-resc:
		if res.Err != "" {
			return nil, errors.New(res.Err)
		}
		return res, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *ProgCache) writeToChild(req *ProgRequest, resc chan<- *ProgResponse) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.nextID++
	req.ID = c.nextID
	c.inFlight[req.ID] = resc
	defer func() {
		if err != nil {
			delete(c.inFlight, req.ID)
		}
	}()
	if err := c.jenc.Encode(req); err != nil {
		return err
	}
	if err := c.bw.WriteByte('\n'); err != nil {
		return err
	}
	if req.Body != nil && req.BodySize > 0 {
		if err := c.bw.WriteByte('"'); err != nil {
			return err
		}
		e := base64.NewEncoder(base64.StdEncoding, c.bw)
		wrote, err := io.Copy(e, req.Body)
		if err != nil {
			return err
		}
		if err := e.Close(); err != nil {
			return nil
		}
		if wrote != req.BodySize {
			return errors.New("short write")
		}
		if _, err := c.bw.WriteString("\"\n"); err != nil {
			return err
		}
	}
	if err := c.bw.Flush(); err != nil {
		return err
	}
	return nil
}

func (c *ProgCache) get(a ActionID, wantFile bool) (_ Entry, path string, _ error) {
	if !c.can[cmdGet] {
		// They can't do a "get". Maybe they're a write-only cache.
		return Entry{}, "", &entryNotFoundError{}
	}
	ctx := context.Background() // TODO(bradfitz): change Cache interface to pass one? add timeout?
	res, err := c.send(ctx, &ProgRequest{
		Command:      cmdGet,
		WantDiskPath: wantFile,
		ActionID:     a[:],
	})
	if err != nil {
		return Entry{}, "", &entryNotFoundError{Err: err}
	}
	if res.Miss {
		return Entry{}, "", &entryNotFoundError{}
	}
	e := Entry{
		Size: res.Size,
		Time: time.Unix(0, res.TimeNanos),
	}
	if copy(e.OutputID[:], res.OutputID) != len(res.OutputID) {
		return Entry{}, "", &entryNotFoundError{errors.New("incomplete ProgResponse OutputID")}
	}
	return e, res.DiskPath, nil
}

func (c *ProgCache) Get(a ActionID) (Entry, error) {
	e, _, err := c.get(a, false)
	if err != nil {
		log.Printf("XXX Cache.Get(%v) = %v", a, err)
		return Entry{}, err
	}
	log.Printf("XXX Cache.Get(%v) = %+v", a, e)
	return e, nil
}

func (c *ProgCache) GetFile(a ActionID) (file string, _ Entry, _ error) {
	e, file, err := c.get(a, true)
	if err != nil {
		return "", Entry{}, err
	}
	return file, e, nil
}

func (c *ProgCache) GetBytes(a ActionID) ([]byte, Entry, error) {
	e, file, err := c.get(a, true)
	if err != nil {
		return nil, Entry{}, err
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, Entry{}, &entryNotFoundError{Err: err}
	}
	if sha256.Sum256(data) != e.OutputID {
		return nil, Entry{}, &entryNotFoundError{Err: errors.New("bad checksum")}
	}
	return data, e, nil
}

func (c *ProgCache) Trim() error {
	if !c.can[cmdTrim] {
		return nil
	}
	ctx := context.Background() // TODO(bradfitz): change Cache interface to pass one? add timeout?
	_, err := c.send(ctx, &ProgRequest{Command: cmdTrim})
	return err
}

func (c *ProgCache) GetMmap(a ActionID) ([]byte, Entry, error) {
	e, file, err := c.get(a, true)
	if err != nil {
		return nil, Entry{}, err
	}
	md, err := mmap.Mmap(file)
	if err != nil {
		return nil, Entry{}, err
	}
	if int64(len(md.Data)) != e.Size {
		return nil, Entry{}, &entryNotFoundError{Err: errors.New("file incomplete")}
	}
	return md.Data, e, nil
}

func (c *ProgCache) OutputFile(o OutputID) string {
	ctx := context.Background() // TODO(bradfitz): change Cache interface to pass one? add timeout?

	if !c.can[cmdOutputFile] {
		// TODO(bradfitz): what? figure out how OutputFile is even used. Maybe remove/change it.
		base.Fatalf("GOCACHEPROG doesn't support output-file")
	}

	res, err := c.send(ctx, &ProgRequest{
		Command:  cmdOutputFile,
		ObjectID: o[:],
	})
	if err != nil {
		panic(err) // TODO(bradfitz): add error return to OutputFile? document OutputFile. when's it used?
	}
	if res.DiskPath == "" {
		base.Fatalf("no DiskPath returned from GOCACHEPROG output-file call for %x", o)
	}
	return res.DiskPath
}

func (c *ProgCache) PutBytes(a ActionID, b []byte) error {
	_, _, err := c.Put(a, bytes.NewReader(b))
	return err
}

func (c *ProgCache) Put(a ActionID, file io.ReadSeeker) (OutputID, int64, error) {
	return c.put(a, file, true)
}

func (c *ProgCache) PutNoVerify(a ActionID, file io.ReadSeeker) (OutputID, int64, error) {
	return c.put(a, file, false)
}

func (c *ProgCache) put(a ActionID, file io.ReadSeeker, allowVerify bool) (OutputID, int64, error) {
	// TODO(bradfitz): allowVerify mode

	// Compute output ID.
	h := sha256.New()
	if _, err := file.Seek(0, 0); err != nil {
		return OutputID{}, 0, err
	}
	size, err := io.Copy(h, file)
	if err != nil {
		return OutputID{}, 0, err
	}
	var out OutputID
	h.Sum(out[:0])

	if _, err := file.Seek(0, 0); err != nil {
		return OutputID{}, 0, err
	}

	if !c.can[cmdPut] {
		// Child is a read-only cache. Do nothing.
		return out, size, nil
	}

	ctx := context.Background() // TODO(bradfitz): change Cache interface to pass one? add timeout?
	_, err = c.send(ctx, &ProgRequest{
		Command:  cmdPut,
		ActionID: a[:],
		ObjectID: out[:],
		Body:     file,
		BodySize: size,
	})
	return out, size, err
}

func (c *ProgCache) FuzzDir() string {
	// TODO(bradfitz): figure out what to do here. For now just use the
	// disk-based default.
	return c.fuzzDirCache.FuzzDir()
}
