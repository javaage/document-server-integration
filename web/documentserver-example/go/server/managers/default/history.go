/**
 *
 * (c) Copyright Ascensio System SIA 2023
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
package dmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/ONLYOFFICE/document-server-integration/config"
	"github.com/ONLYOFFICE/document-server-integration/server/managers"
	"github.com/ONLYOFFICE/document-server-integration/server/models"
	"github.com/ONLYOFFICE/document-server-integration/server/shared"
	"github.com/ONLYOFFICE/document-server-integration/utils"
	"go.uber.org/zap"
)

type DefaultHistoryManager struct {
	logger *zap.SugaredLogger
	managers.StorageManager
	managers.JwtManager
	config config.ApplicationConfig
}

func NewDefaultHistoryManager(logger *zap.SugaredLogger, sm managers.StorageManager,
	jwt managers.JwtManager, config config.ApplicationConfig) managers.HistoryManager {
	return &DefaultHistoryManager{
		logger,
		sm,
		jwt,
		config,
	}
}

func (hm DefaultHistoryManager) readHistoryChanges(cpath string) ([]models.Changes, error) {
	var mchanges []models.Changes
	changes, err := hm.StorageManager.ReadFile(cpath)
	if err != nil {
		return mchanges, err
	}

	if err := json.Unmarshal(changes, &mchanges); err != nil {
		return mchanges, err
	}

	return mchanges, nil
}

func (hm DefaultHistoryManager) readHistoryFileKey(keyPath string) (string, error) {
	key, err := hm.StorageManager.ReadFile(keyPath)
	if err != nil {
		return "", err
	}

	return string(key[:]), nil
}

func (hm DefaultHistoryManager) buildNextHistory(mchanges []models.Changes, key string, version int) models.History {
	return models.History{
		Changes: mchanges,
		Key:     key,
		Created: mchanges[len(mchanges)-1].Created,
		User:    mchanges[len(mchanges)-1].User,
		Version: version,
	}
}

func (hm DefaultHistoryManager) signHistorySet(set *managers.HistorySet) error {
	var err error
	if hm.config.JwtSecret != "" && hm.config.JwtEnabled {
		set.Token, err = hm.JwtManager.JwtSign(set, []byte(hm.config.JwtSecret))
		if err != nil {
			return err
		}
	}

	return nil
}

func (hm DefaultHistoryManager) fetchNextHistoryEntry(remoteAddress string, filename string, version int) (models.History, managers.HistorySet, error) {
	var (
		hresp  models.History
		hsresp managers.HistorySet
	)

	storagePath, err := hm.StorageManager.GetRootFolder(remoteAddress)
	if err != nil {
		return hresp, hsresp, err
	}

	histPath := path.Join(storagePath, filename+shared.ONLYOFFICE_HISTORY_POSTFIX, fmt.Sprint(version))
	mchanges, err := hm.readHistoryChanges(path.Join(histPath, "changes.json"))
	if err != nil {
		return hresp, hsresp, err
	}

	key, err := hm.readHistoryFileKey(path.Join(histPath, "key.txt"))
	if err != nil {
		return hresp, hsresp, err
	}

	var hset managers.HistorySet
	url := hm.StorageManager.GeneratePublicFileUri(filename, managers.FileMeta{
		Version:         version,
		DestinationPath: "prev" + utils.GetFileExt(filename),
	})

	changesUrl := hm.StorageManager.GeneratePublicFileUri(filename, managers.FileMeta{
		Version:         version,
		DestinationPath: "diff.zip",
	})

	if version > 1 {
		prevHistPath := path.Join(storagePath, filename+shared.ONLYOFFICE_HISTORY_POSTFIX, fmt.Sprint(version-1))
		prevKey, err := hm.readHistoryFileKey(path.Join(prevHistPath, "key.txt"))
		if err != nil {
			return hresp, hsresp, err
		}

		prevUrl := hm.StorageManager.GeneratePublicFileUri(filename, managers.FileMeta{
			Version:         version - 1,
			DestinationPath: "prev" + utils.GetFileExt(filename),
		})

		hset = managers.HistorySet{
			ChangesUrl: changesUrl,
			Key:        key,
			Url:        url,
			Version:    version,
			Previous: managers.HistoryPrevious{
				Key: prevKey,
				Url: prevUrl,
			},
		}
	} else {
		hset = managers.HistorySet{
			ChangesUrl: changesUrl,
			Key:        key,
			Url:        url,
			Version:    version,
		}
	}

	if err := hm.signHistorySet(&hset); err != nil {
		return hresp, hset, err
	}

	return hm.buildNextHistory(mchanges, key, version), hset, nil
}

func (hm DefaultHistoryManager) GetHistory(filename string, remoteAddress string) (managers.HistoryRefresh, []managers.HistorySet, error) {
	var (
		version int = 1
		rhist   managers.HistoryRefresh
		setHist []managers.HistorySet
	)

	rootPath, err := hm.StorageManager.GetRootFolder(remoteAddress)
	if err != nil {
		return rhist, setHist, err
	}

	for {
		hpath := path.Join(rootPath, filename+shared.ONLYOFFICE_HISTORY_POSTFIX, fmt.Sprint(version))
		if hm.StorageManager.PathExists(hpath) {
			hist, histSet, err := hm.fetchNextHistoryEntry(remoteAddress, filename, version)
			if err != nil {
				return rhist, setHist, err
			}

			rhist.History = append(rhist.History, hist)
			setHist = append(setHist, histSet)
			version += 1
		} else {
			break
		}
	}

	rhist.CurrentVersion = fmt.Sprint(version)
	currMeta, err := hm.readHistoryChanges(path.Join(rootPath, filename+shared.ONLYOFFICE_HISTORY_POSTFIX, filename+".json"))
	if err != nil {
		return rhist, setHist, err
	}

	docKey, err := hm.StorageManager.GenerateFileHash(filename, remoteAddress)
	if err != nil {
		return rhist, setHist, err
	}

	rhist.History = append(rhist.History, models.History{
		Changes: currMeta,
		User:    currMeta[len(currMeta)-1].User,
		Created: currMeta[len(currMeta)-1].Created,
		Key:     docKey,
		Version: version,
	})

	currSet := managers.HistorySet{
		Key:     docKey,
		Url:     hm.StorageManager.GeneratePublicFileUri(filename, managers.FileMeta{}),
		Version: version,
	}

	if err := hm.signHistorySet(&currSet); err != nil {
		return rhist, setHist, err
	}

	setHist = append(setHist, currSet)
	return rhist, setHist, nil
}

func (hm DefaultHistoryManager) CreateMeta(filename string, remoteAddress string, changes []models.Changes) error {
	rootPath, err := hm.StorageManager.GetRootFolder(remoteAddress)
	if err != nil {
		return err
	}

	hpath := path.Join(rootPath, filename+shared.ONLYOFFICE_HISTORY_POSTFIX)
	bdata, err := json.MarshalIndent(changes, " ", "")
	if err != nil {
		return err
	}

	if err := hm.StorageManager.CreateDirectory(hpath); err != nil {
		return err
	}

	return hm.StorageManager.CreateFile(bytes.NewReader(bdata), path.Join(hpath, filename+".json"))
}

func (hm DefaultHistoryManager) isMeta(filename string, remoteAddress string) bool {
	rootPath, err := hm.StorageManager.GetRootFolder(remoteAddress)
	if err != nil {
		return false
	}

	hpath := path.Join(rootPath, filename+shared.ONLYOFFICE_HISTORY_POSTFIX)
	return hm.StorageManager.PathExists(path.Join(hpath, filename+".json"))
}

func (hm DefaultHistoryManager) CreateHistory(cbody models.Callback, remoteAddress string) error {
	var version int = 1
	spath, err := hm.StorageManager.GetRootFolder(remoteAddress)
	if err != nil {
		return err
	}

	prevFilePath, err := hm.StorageManager.GenerateFilePath(cbody.Filename, remoteAddress)
	if err != nil {
		return err
	}

	hdir := path.Join(spath, cbody.Filename+shared.ONLYOFFICE_HISTORY_POSTFIX)
	if !hm.isMeta(cbody.Filename, remoteAddress) {
		return fmt.Errorf("file %s no longer exists", cbody.Filename)
	}

	for {
		histDirVersion := path.Join(hdir, fmt.Sprint(version))
		if !hm.StorageManager.PathExists(histDirVersion) {
			hm.StorageManager.CreateDirectory(histDirVersion)

			hm.StorageManager.MoveFile(path.Join(hdir, cbody.Filename+".json"), path.Join(histDirVersion, "changes.json"))

			cbytes, err := json.Marshal(cbody.History.Changes)
			if err != nil {
				return err
			}

			hm.StorageManager.CreateFile(bytes.NewReader(cbytes), path.Join(hdir, cbody.Filename+".json"))
			hm.StorageManager.CreateFile(bytes.NewReader([]byte(cbody.Key)), path.Join(histDirVersion, "key.txt"))
			hm.StorageManager.MoveFile(prevFilePath, path.Join(histDirVersion, "prev"+utils.GetFileExt(cbody.Filename)))
			resp, err := http.Get(cbody.ChangesUrl)
			if err != nil {
				return err
			}

			defer resp.Body.Close()
			hm.StorageManager.CreateFile(resp.Body, path.Join(histDirVersion, "diff.zip"))
			break
		}
		version += 1
	}

	return nil
}