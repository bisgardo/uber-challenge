package types

type Movie struct {
	Title             string
	Locations         []Location
	Actors            []string
	Director          string
	Distributor       string
	Writer            string
	ProductionCompany string
	ReleaseYear       int
}

type Location struct {
	Name        string
	FunFact     string
	Coordinates Coordinates
}

type Coordinates struct {
	Lat float32
	Lng float32
}

type IdMoviePair struct {
	Id    int64
	Movie Movie
}

// Comparator for sorting movie list.
type ByTitle []IdMoviePair

func (ms ByTitle) Len() int {
	return len(ms)
}
func (ms ByTitle) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}
func (ms ByTitle) Less(i, j int) bool {
	return ms[i].Movie.Title < ms[j].Movie.Title
}
