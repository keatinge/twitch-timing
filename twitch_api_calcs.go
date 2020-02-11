package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"time"
)

const client_id = "" // TODO: Read from env variable
const client_secret = "" // TODO: Read from env variable

const N_BINS = 100
const N_VIDEOS = 100


type UsersResponse struct {
	Id string `json:"id"`
	Login string `json:"login"`
}

func ApiReq(req_url string, params map[string]string) []byte {
	values := make(url.Values)
	for k, v := range params {
		values.Set(k, v)
	}
	req, err := http.NewRequest("GET", req_url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.URL.RawQuery = values.Encode()
	req.Header.Set("Client-ID", client_id)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client_secret))
	fmt.Println("Making request to", req.URL.String())


	client := &http.Client{}
	resp, err := client.Do(req)
	fmt.Println(resp.Status)

	if err != nil {
		log.Fatal(err)
	}

	body, read_err := ioutil.ReadAll(resp.Body)
	if read_err != nil {
		log.Fatal(read_err)
	}
	return body
}

func get_user_id(username string) string {
	var params = map[string]string{
		"login": username,
	}

	body := ApiReq("https://api.twitch.tv/helix/users", params)
	api_res := make(map[string][]UsersResponse)
	err := json.Unmarshal(body, &api_res)
	if err != nil {
		log.Fatal(err)
	}
	if len(api_res["data"]) <= 0 {
		return ""
	}
	return api_res["data"][0].Id
}


type Video struct {
	Published_at time.Time
	Id string
	Duration string
	Title string
	Type string
}

type PaginationStruct struct {
	Cursor string
}
type VideoResponse struct {
	Data []Video
	Pagination PaginationStruct
}



func get_vods(user_id string) []Video {
	var ret_videos []Video
	after := ""
	for len(ret_videos) < N_VIDEOS {
		fmt.Println(len(ret_videos))
		params := map[string]string {
			"user_id": user_id,
			"after": after,
		}
		body := ApiReq("https://api.twitch.tv/helix/videos", params)
		vr := VideoResponse{}
		err := json.Unmarshal(body, &vr)
		if err != nil {
			log.Fatal(err)
		}

		for _, video := range vr.Data {
			if video.Type == "archive" {
				ret_videos = append(ret_videos, video)
			}

		}

		if vr.Pagination.Cursor == after || vr.Pagination.Cursor == "" {
			break
		}
		after = vr.Pagination.Cursor
	}

	return ret_videos
}

func get_bins(stream_start time.Time, stream_end time.Time) []int64 {
	n_splits := N_BINS
	bins := make([]int64, n_splits)


	for i, _ := range bins {
		bin_time := (0.5 + float64(i)) * 24.0/float64(n_splits) * 60 * 60

		d1 := time.Date(stream_start.Year(), stream_start.Month(), stream_start.Day(), 0, 0, int(bin_time), 0, time.UTC)
		d2 := time.Date(stream_end.Year(), stream_end.Month(), stream_end.Day(), 0, 0, int(bin_time), 0, time.UTC)

		d1_in := stream_start.Before(d1) && stream_end.After(d1)
		d2_in := stream_start.Before(d2) && stream_end.After(d2)

		if d1_in || d2_in {
			bins[i] += 1
		}

	}
	return bins
}

func get_end_time(vod Video) time.Time {
	dur, err := time.ParseDuration(vod.Duration)
	if err != nil {
		log.Fatal(err)
	}
	end_time := vod.Published_at.Add(dur)
	return end_time
}

func get_bin_sums(vods []Video) []int64 {
	bin_sum := make([]int64, N_BINS)
	for _, vod := range vods {
		//t_format := "3:04PM MST 2006-01-02"
		end_time := get_end_time(vod)
		bins := get_bins(vod.Published_at, end_time)
		for i, x := range bins {
			bin_sum[i] += x
		}

	}
	return bin_sum
}

func get_dow_bin_sum(vods []Video) []int64 {
	bins := make([]int64, 7)
	loc, err := time.LoadLocation("US/Eastern")
	if err != nil {
		log.Println("Illegal loc")
		return nil
	}
	for _, vod := range vods {
		start := vod.Published_at.In(loc)
		end := get_end_time(vod).In(loc)
		s_day := start.Weekday()
		e_day := end.Weekday()

		bins[s_day] += 1
		if s_day != e_day {
			bins[e_day] += 1
		}
	}
	return bins
}

func get_stream_dirs(vods []Video) []float64 {
	durs := make([]float64, len(vods))

	for i, vod := range vods {
		dur, err := time.ParseDuration(vod.Duration)
		if err != nil {
			log.Println("Couldn't parse duration", err)
		}
		durs[i] = dur.Hours()
	}
	return durs
}

func get_bin_timings(n_bins int) []float64 {
	ret := make([]float64, n_bins)
	for i := 0; i < len(ret); i++ {
		actual := (0.5 + float64(i))/float64(n_bins) * 24.0
		ten_k := 10_000.0
		ret[i] = math.Round(ten_k*actual) / ten_k
	}
	return ret
}

func filter_short(vods []Video) []Video {
	var ret []Video
	discarded := 0

	th, _ := time.ParseDuration("20m")
	for _, vod := range vods {
		dur, err := time.ParseDuration(vod.Duration)
		if err != nil {
			log.Println("Couldn't parse duration", err)
		}
		if dur > th {
			ret = append(ret, vod)
		} else {
			discarded += 1
		}
	}
	fmt.Println("Discarded", discarded)
	return ret
}

func get_bin_sum_and_timings(user string) (int, []int64, []float64, []int64, []float64) {
	user_id := get_user_id(user)
	vods := get_vods(user_id)
	vods_real := filter_short(vods)
	bin_sums := get_bin_sums(vods_real)
	bin_timings := get_bin_timings(len(bin_sums))
	dow_bin_sum := get_dow_bin_sum(vods_real)
	durs := get_stream_dirs(vods_real)
	return len(vods_real), bin_sums, bin_timings, dow_bin_sum, durs
}



