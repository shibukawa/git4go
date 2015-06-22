package git4go

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type OdbBackendPacked struct {
	OdbBackendBase
	packFolder string
	packs      []*PackFile
	lastFound  *PackFile
}

func NewOdbBackendPacked(objectsDir string) *OdbBackendPacked {
	folderPath := filepath.Join(objectsDir, "pack")
	info, err := os.Stat(folderPath)
	if os.IsNotExist(err) || !info.IsDir() {
		return nil
	}
	result := &OdbBackendPacked{
		packFolder: folderPath,
	}
	result.Refresh()
	return result
}

func (o *OdbBackendPacked) Read(oid *Oid) (*OdbObject, error) {
	entry, err := o.findEntry(oid)
	if err != nil {
		return nil, err
	}
	obj, _, err := entry.PackFile.unpack(entry.Offset)
	return obj, err
}

func (o *OdbBackendPacked) ReadPrefix(oid *Oid, length int) (*Oid, *OdbObject, error) {
	return nil, nil, errors.New("not implemented")
}

func (o *OdbBackendPacked) ReadHeader(oid *Oid) (ObjectType, uint64, error) {
	entry, err := o.findEntry(oid)
	if err != nil {
		return ObjectBad, 0, err
	}
	objType, size, err := entry.PackFile.resolveHeader(entry.Offset)
	return objType, size, err
}

func (o *OdbBackendPacked) Write(data []byte, objType ObjectType) (*Oid, error) {
	return nil, errors.New("not implemented")
}

func (o *OdbBackendPacked) Exists(oid *Oid) bool {
	_, err := o.findEntry(oid)
	return err == nil
}

func (o *OdbBackendPacked) ExistsPrefix(shortOid *Oid, length int) (*Oid, error) {
	entry, err := o.findEntryByPrefix(shortOid, length)
	if err != nil {
		return nil, err
	} else {
		return entry.Sha1, err
	}
}

func (o *OdbBackendPacked) Refresh() error {
	dir, err := os.Open(o.packFolder)
	defer dir.Close()
	if err != nil {
		return errors.New("failed to refresh packfiles")
	}
	stat, err := dir.Stat()
	if !stat.IsDir() || err != nil {
		return errors.New("failed to refresh packfiles")
	}
	names, err := dir.Readdirnames(0)
	if err != nil {
		return errors.New("failed to refresh packfiles")
	}
	for _, name := range names {
		if !strings.HasSuffix(name, ".idx") {
			continue
		}
		path := filepath.Join(o.packFolder, name)
		baseName := path[:len(path)-4]
		found := false
		for _, packFile := range o.packs {
			if packFile.baseName == baseName {
				found = true
				break
			}
		}
		if found {
			continue
		}
		pack, err := GetPack(path)
		if err == nil {
			o.packs = append(o.packs, pack)
		}
	}
	return nil
}

// internal functions

func (o *OdbBackendPacked) findEntry(oid *Oid) (*PackEntry, error) {
	entry, notFound, err := o.findEntryInternal(oid)
	if err == nil {
		return entry, nil
	}
	if notFound {
		err = o.Refresh()
		if err != nil {
			return nil, err
		}
	}
	entry, _, err = o.findEntryInternal(oid)
	return entry, err
}

func (o *OdbBackendPacked) findEntryInternal(oid *Oid) (*PackEntry, bool, error) {
	if o.lastFound != nil {
		entry, notFound, err := o.lastFound.findEntry(oid, GIT_OID_HEXSZ)
		if !notFound && err != nil {
			return nil, false, err
		}
		if err == nil {
			return entry, false, nil
		}
	}
	for _, pack := range o.packs {
		if pack == o.lastFound {
			continue
		}
		entry, notFound, err := pack.findEntry(oid, GIT_OID_HEXSZ)
		if !notFound && err != nil {
			return nil, false, err
		}
		if err == nil {
			o.lastFound = pack
			return entry, false, nil
		}
	}
	return nil, true, errors.New("failed to find pack entry: " + oid.String())
}

func (o *OdbBackendPacked) findEntryByPrefix(shortOid *Oid, length int) (*PackEntry, error) {
	entry, notFound, err := o.findEntryByPrefixInternal(shortOid, length)
	if err == nil {
		return entry, nil
	}
	if notFound {
		err = o.Refresh()
		if err != nil {
			return nil, err
		}
	}
	entry, _, err = o.findEntryByPrefixInternal(shortOid, length)
	return entry, err
}

func (o *OdbBackendPacked) findEntryByPrefixInternal(shortOid *Oid, length int) (*PackEntry, bool, error) {
	var foundEntry *PackEntry = nil
	if o.lastFound != nil {
		entry, notFound, err := o.lastFound.findEntry(shortOid, length)
		if !notFound && err != nil {
			return nil, false, err
		}
		if err == nil {
			foundEntry = entry
		}
	}
	for _, pack := range o.packs {
		entry, notFound, err := pack.findEntry(shortOid, length)
		if !notFound && err != nil {
			return nil, false, err
		}
		if err == nil {
			if foundEntry != nil && !foundEntry.Sha1.Equal(entry.Sha1) {
				return nil, false, errors.New("found multiple pack entries for: " + shortOid.String())
			}
			o.lastFound = pack
		}
	}
	if foundEntry != nil {
		return foundEntry, false, nil
	} else {
		return nil, true, errors.New("failed to find pack entry: " + shortOid.String())
	}
}
