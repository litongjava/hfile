package utils

import "github.com/litongjava/hfile/model"

func CompareForUpload(localFiles, remoteFiles map[string]model.FileMeta) []model.FileMeta {
	var result []model.FileMeta
	for path, l := range localFiles {
		if r, ok := remoteFiles[path]; ok {
			if l.Hash != r.Hash && l.ModTime > r.ModTime {
				result = append(result, l)
			}
		} else {
			result = append(result, l)
		}
	}
	return result
}

func CompareForDownload(local, remote map[string]model.FileMeta) []model.FileMeta {
	var result []model.FileMeta
	for path, r := range remote {
		if l, ok := local[path]; ok {
			if r.Hash != l.Hash && r.ModTime > l.ModTime {
				result = append(result, r)
			}
		} else {
			result = append(result, r)
		}
	}
	return result
}
