package main

type GetVideosResp struct {
	Total     int                 `json:"total"`
	Videos    []GetVideosRespElem `json:"videos,omitempty"`
	NextToken string              `json:"next_token,omitempty"`
	Time      string              `json:"time"`
}

type GetVideosRespElem struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Size string `json:"size,omitempty"`
}

type GetImagesResp struct {
	Total     int                 `json:"total"`
	Images    []GetImagesRespElem `json:"images,omitempty"`
	NextToken string              `json:"next_token,omitempty"`
	Time      string              `json:"time"`
}

type GetImagesRespElem struct {
	Elems []GetImagesRespElemElem `json:"elems"`
}

type GetImagesRespElemElem struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Size string `json:"size,omitempty"`
}
