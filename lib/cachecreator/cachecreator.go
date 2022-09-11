package cachecreator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yayuyokitano/eggshellver/lib/logging"
	"github.com/yayuyokitano/eggshellver/lib/queries"
)

func appendSongData(v *map[string]queries.SongData, s []queries.SongData) {
	for _, song := range s {
		(*v)[song.MusicID] = song
	}
}

func appendArtistData(v *map[string]queries.ArtistData, a []queries.SongData) {
	for _, artist := range a {
		(*v)[artist.ArtistData.ArtistName] = artist.ArtistData
	}
}

func printProgress(i int, total int) {
	fmt.Printf("%d/%d\n", i, total)
}

func songMapToSlice(m map[string]queries.SongData) (s []queries.SongData) {
	for _, v := range m {
		s = append(s, v)
	}
	return
}

func artistMapToSlice(m map[string]queries.ArtistData) (u []queries.UserStub) {
	for _, v := range m {
		u = append(u, queries.UserStub{
			UserID:         v.ArtistID,
			EggsID:         v.ArtistName,
			DisplayName:    v.DisplayName,
			IsArtist:       true,
			ImageDataPath:  v.ImageDataPath,
			PrefectureCode: v.PrefectureCode,
			ProfileText:    v.Profile,
		})
	}
	return
}

func completeCache(songs []queries.SongData, artists []queries.UserStub) (err error) {
	fmt.Println("Writing artists to DB...")
	inserted, _, err := queries.PostUserStubs(context.Background(), artists)
	if err != nil {
		logging.FailCache(err)
		return
	}
	logging.AddCachedUsers(int(inserted))

	time.Sleep(5 * time.Second) //a little wait because it otherwise seems to be a bit unreliable
	fmt.Println("Writing songs to DB...")
	n, err := queries.PostSongs(context.Background(), songs)
	if err != nil {
		logging.FailCache(err)
		return
	}
	logging.AddSongs(int(n))
	logging.CompleteCache()
	return
}

func runFullCache() {
	time.Sleep(5 * time.Second)
	fmt.Println("Creating cache...")
	initialResp := getSongs(0, 1000)
	total := initialResp.TotalCount
	cur := 990
	printProgress(cur, total)
	songMap := make(map[string]queries.SongData)
	artistMap := make(map[string]queries.ArtistData)

	appendSongData(&songMap, initialResp.Data)
	appendArtistData(&artistMap, initialResp.Data)

	for cur < total {
		time.Sleep(1 * time.Second)
		resp := getSongs(cur, 1000)
		total = resp.TotalCount
		appendSongData(&songMap, resp.Data)
		cur += 990 //safety margin

		printProgress(cur, total)
		appendArtistData(&artistMap, resp.Data)
	}
	fmt.Println("Transforming Data...")

	songs := songMapToSlice(songMap)
	artists := artistMapToSlice(artistMap)

	err := completeCache(songs, artists)
	if err != nil {
		panic(err)
	}

	fmt.Println("Done!")
}

func AttemptRunPartialCache() {
	songResp, err := getRecentSongs()
	if err != nil {
		return
	}
	exists, err := queries.SongExists(context.Background(), songResp.Data[len(songResp.Data)-1].MusicID)
	if err != nil {
		log.Println(err)
		return
	}
	if !exists {
		runFullCache()
		return
	}
	songMap := make(map[string]queries.SongData)
	artistMap := make(map[string]queries.ArtistData)

	appendSongData(&songMap, songResp.Data)
	appendArtistData(&artistMap, songResp.Data)

	songs := songMapToSlice(songMap)
	artists := artistMapToSlice(artistMap)

	completeCache(songs, artists)
}

func StartCacheLoop(t time.Duration) {
	fmt.Println("Starting cache loop...")
	for {
		AttemptRunPartialCache()
		time.Sleep(t)
	}
}
