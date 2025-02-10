package scoring

import "time"

const (
	Dir        = "data"
	NumWorkers = 5
)

type ScoredRow struct {
	Key     CompanyYearKey
	Metrics map[string]float64
}

// Wrap the result in the channel
type RowResult struct {
	Row ScoredRow
	Err error
}

type CompanyYearKey struct {
	CompanyID string
	Year      int
}

type rowData struct {
	Date    time.Time
	Numeric map[string]float64
}

// just to test the json file
type RawJSONRow struct {
	CompanyID string   `json:"company_id"`
	DateStr   string   `json:"date"`
	Dis1      *float64 `json:"dis_1,omitempty"`
	Dis2      *float64 `json:"dis_2,omitempty"`
	Dis3      *float64 `json:"dis_3,omitempty"`
	Dis4      *float64 `json:"dis_4,omitempty"`
}

type LoaderRegistry struct {
	registry map[string]DataLoader
}

// GetLoader returns the DataLoader for a given file extension, if found.
func (lr *LoaderRegistry) GetLoader(ext string) (DataLoader, bool) {
	loader, ok := lr.registry[ext]
	return loader, ok
}

// RegisterLoader lets you add or overwrite a DataLoader for a specific extension.
func (lr *LoaderRegistry) RegisterLoader(ext string, loader DataLoader) {
	lr.registry[ext] = loader
}

// NewLoaderRegistry initializes a default registry with a CSV loader.
func NewLoaderRegistry() *LoaderRegistry {
	return &LoaderRegistry{
		registry: map[string]DataLoader{
			".csv":  CSVLoader{},
			".json": JSONLoader{},
			// ".sql": RepoLoader{ DB: *pgpool },
		},
	}
}

// DataLoaderService orchestrates reading multiple files from a directory
// using the loaders from the LoaderRegistry.
// Load all files in directory
type DataLoaderService struct {
	registry *LoaderRegistry
}

func NewDataLoaderService(lr *LoaderRegistry) *DataLoaderService {
	return &DataLoaderService{registry: lr}
}

var DatasetKeys = map[string]string{
	"disclosure": "disclosure_data",
	"waste":      "waste_data",
	"emissions":  "emissions_data",
}
