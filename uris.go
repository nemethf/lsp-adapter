package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/sourcegraph/go-langserver/pkg/lsp"

	"github.com/pkg/errors"
)

func (p *cloneProxy) cloneWorkspaceToCache(globs []string) error {
	fs := &remoteFS{conn: p.client, traceID: p.sessionID.String()}
	err := fs.Clone(p.ctx, p.workspaceCacheDir(), globs)
	if err != nil {
		return errors.Wrap(err, "failed to clone workspace to local cache")
	}

	log.Printf("Cloned workspace to %s", p.workspaceCacheDir())
	return nil
}

func (p *cloneProxy) cleanWorkspaceCache() error {
	log.Printf("Removing workspace cache from %s", p.workspaceCacheDir())
	return os.RemoveAll(p.workspaceCacheDir())
}

func (p *cloneProxy) workspaceCacheDir() string {
	return filepath.Join(*cacheDir, p.sessionID.String())
}

func clientToServerURI(uri lsp.DocumentURI, sysCacheDir string) lsp.DocumentURI {
	// sysCacheDir needs to be converted from a local path to a URI path
	cacheDir := filepath.ToSlash(sysCacheDir)

	parsedURI, err := url.Parse(string(uri))

	if err != nil {
		log.Println(fmt.Sprintf("clientToServerURI: err when trying to parse uri %s", uri), err)
		return uri
	}

	if !probablyFileURI(parsedURI) {
		return uri
	}

	// We assume that any path provided by the client to the server
	// is a project path that is relative to '/'
	parsedURI.Path = path.Join(cacheDir, parsedURI.Path)
	return lsp.DocumentURI(parsedURI.String())
}

func serverToClientURI(uri lsp.DocumentURI, sysCacheDir string) lsp.DocumentURI {
	// sysCacheDir needs to be converted from a local path to a URI path
	cacheDir := filepath.ToSlash(sysCacheDir)

	parsedURI, err := url.Parse(string(uri))

	if err != nil {
		log.Println(fmt.Sprintf("serverToClientURI: err when trying to parse uri %s", uri), err)
		return uri
	}

	if !probablyFileURI(parsedURI) {
		return uri
	}

	// Only rewrite uris that point to a location in the workspace cache. If it does
	// point to a cache location, then we assume that the path points to a location in the
	// project.
	if pathHasPrefix(parsedURI.Path, cacheDir) {
		parsedURI.Path = path.Join("/", pathTrimPrefix(parsedURI.Path, cacheDir))
	}

	return lsp.DocumentURI(parsedURI.String())
}

func probablyFileURI(candidate *url.URL) bool {
	if !(candidate.Scheme == "" || candidate.Scheme == "file") {
		return false
	}

	if candidate.Path == "" {
		return false
	}

	return true
}

func pathHasPrefix(s, prefix string) bool {
	return rawHasPrefix(s, prefix, "/")
}

func filepathHasPrefix(s, prefix string) bool {
	return rawHasPrefix(s, prefix, string(os.PathSeparator))
}

// adapted from sourcegraph/go-langserver/util.go
func rawHasPrefix(s, prefix, pathSep string) bool {
	var prefixSlash string
	if prefix != "" && !strings.HasSuffix(prefix, pathSep) {
		prefixSlash = prefix + pathSep
	}
	return s == prefix || strings.HasPrefix(s, prefixSlash)
}

func pathTrimPrefix(s, prefix string) string {
	return rawTrimPrefix(s, prefix, "/")
}

func filepathTrimPrefix(s, prefix string) string {
	return rawTrimPrefix(s, prefix, string(os.PathSeparator))
}

// adapted from sourcegraph/go-langserver/util.go
func rawTrimPrefix(s, prefix, pathSep string) string {
	if s == prefix {
		return ""
	}
	if !strings.HasSuffix(prefix, pathSep) {
		prefix += pathSep
	}
	return strings.TrimPrefix(s, prefix)
}

// WalkURIFields walks the LSP params/result object for fields
// containing document URIs.
//
// If update is non-nil, it updates all document URIs in an LSP
// params/result with the value of f(existingURI). Callers can use
// this to rewrite paths in the params/result.
func WalkURIFields(o interface{}, update func(lsp.DocumentURI) lsp.DocumentURI) {
	var walk func(o interface{}, parent string)
	walk = func(o interface{}, parent string) {
		switch o := o.(type) {
		case map[string]interface{}:
			for k, v := range o { // Location, TextDocumentIdentifier, TextDocumentItem, etc.
				// Handling "rootPath" and "rootUri" special cases the initialize method.
				if k == "uri" || k == "rootPath" || k == "rootUri" || k == "url"{
					s, ok := v.(string)
					if !ok {
						s2, ok2 := v.(lsp.DocumentURI)
						s = string(s2)
						ok = ok2
					}
					if ok {
						if update != nil {
							o[k] = update(lsp.DocumentURI(s))
						}
						continue
					}
				}
				if parent == "changes" {
					new_uri := update(lsp.DocumentURI(k))
					delete(o, k)
					o[string(new_uri)] = v
				}
				walk(v, k)
			}
		case []interface{}: // Location[]
			for k, v := range o {
				walk(v, string(k))
			}
		default: // structs with a "URI" field
			rv := reflect.ValueOf(o)
			if rv.Kind() == reflect.Ptr {
				rv = rv.Elem()
			}
			if rv.Kind() == reflect.Struct {
				if fv := rv.FieldByName("URI"); fv.Kind() == reflect.String {
					if update != nil {
						fv.SetString(string(update(lsp.DocumentURI(fv.String()))))
					}
				}
				for i := 0; i < rv.NumField(); i++ {
					fv := rv.Field(i)
					if fv.Kind() == reflect.Ptr || fv.Kind() == reflect.Struct || fv.Kind() == reflect.Array {
						walk(fv.Interface(), "n/a")
					}
				}
			}
		}
	}
	walk(o, "top")
}
