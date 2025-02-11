package scoring

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"go.uber.org/zap"

	c "esgbook-software-engineer-technical-test-2024/pkg/config"
)

// buildDependencyGraph list of metrics that depend on 'a'
func buildDependencyGraph(logger *zap.Logger, scoreConfig *c.Config) (map[string][]string, map[string]int) {
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	for _, m := range scoreConfig.Metrics {
		inDegree[m.Name] = 0
		graph[m.Name] = []string{}
	}

	for _, m := range scoreConfig.Metrics {
		for _, param := range m.Operation.Parameters {
			if strings.HasPrefix(param.Source, "self.") {
				depMetric := strings.TrimPrefix(param.Source, "self.")
				graph[depMetric] = append(graph[depMetric], m.Name)
				inDegree[m.Name]++
			}
		}
	}

	logger.Sugar().Debug("No value for key",
		zap.Any("order", graph),
		zap.Any("order", inDegree))

	return graph, inDegree
}

// topologicalSort check if dependency cycle exists
func topologicalSort(
	logger *zap.Logger,
	scoreConfig *c.Config,
	graph map[string][]string,
	inDegree map[string]int,
) ([]string, error) {
	var queue []string
	for metricName, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, metricName)
		}
	}

	var sorted []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		sorted = append(sorted, current)

		for _, dep := range graph[current] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
	}

	if len(sorted) < len(scoreConfig.Metrics) {
		logger.Error("Cycle detected in metric dependencies")
		return nil, fmt.Errorf("cycle detected in metric dependencies")
	}

	logger.Sugar().Debug("No value for key",
		zap.Any("order", sorted))

	return sorted, nil
}

func getAllDataCompanyKeys(datasets map[string]map[CompanyYearKey]map[string]float64) []CompanyYearKey {
	unique := make(map[CompanyYearKey]bool)
	for _, ds := range datasets {
		for key := range ds {
			unique[key] = true
		}
	}
	out := make([]CompanyYearKey, 0, len(unique))
	for k := range unique {
		out = append(out, k)
	}
	return out
}

func evaluateMetric(
	ctx context.Context,
	logger *zap.Logger,
	metric c.Metric,
	key CompanyYearKey,
	//results map[CompanyYearKey]map[string]float64,
	results map[string]float64,
	datasets map[string]map[CompanyYearKey]map[string]float64,
) (float64, bool) {
	// if we've already computed metric, return it
	if val, ok := results[metric.Name]; ok {
		return val, false
	}

	opFn, ok := Operations[metric.Operation.Type]
	if !ok {
		logger.Sugar().Infow("No value for key",
			zap.String("company_id", key.CompanyID),
			zap.Int("year", key.Year),
			zap.String("operation type", metric.Operation.Type))

		return 0, true
	}

	val, isNull, err := opFn(ctx, logger, metric.Operation.Parameters, key, results, datasets)
	if err != nil {
		logger.Sugar().Infow("No value for key",
			zap.String("company_id", key.CompanyID),
			zap.Int("year", key.Year),
			zap.String("value", fmt.Sprintf("%v", val)))

		return 0, true
	}

	if !isNull {
		results[metric.Name] = val
	}

	return val, isNull
}

// getValue from source file
func getValue(
	logger *zap.Logger,
	source string,
	key CompanyYearKey,
	//results map[CompanyYearKey]map[string]float64,
	results map[string]float64,
	datasets map[string]map[CompanyYearKey]map[string]float64,
) (float64, bool) {
	if strings.HasPrefix(source, "self.") {
		metricName := strings.TrimPrefix(source, "self.")
		val, ok := results[metricName]
		if !ok {
			logger.Sugar().Infow("No value for key",
				zap.String("company_id", key.CompanyID),
				zap.Int("year", key.Year),
				zap.String("value", fmt.Sprintf("%v", val)))
			return 0, true
		}
		return val, false
	}

	parts := strings.Split(source, ".")
	if len(parts) != 2 {
		return 0, true // invalid format => null
	}
	datasetName := parts[0]
	metricKey := parts[1]

	ds, ok := datasets[datasetName]
	if !ok {
		logger.Sugar().Infow("Unknown dataset",
			zap.String("dataset name", datasetName))

		return 0, true
	}
	row, ok := ds[key]
	if !ok {
		logger.Sugar().Infow("no row for key",
			zap.String("company_id", key.CompanyID),
			zap.Int("year", key.Year),
			zap.String("datasetName", datasetName))

		return 0, true
	}
	val, ok := row[metricKey]
	if !ok {
		logger.Sugar().Infow("No value for key",
			zap.String("company_id", key.CompanyID),
			zap.Int("year", key.Year),
			zap.String("datasetName", datasetName),
			zap.String("metricKey", metricKey),
			zap.String("value", fmt.Sprintf("%v", val)))

		return 0, true
	}
	return val, false
}

// parallelComputeScores
func parallelComputeScores(
	ctx context.Context,
	logger *zap.Logger,
	allKeys []CompanyYearKey,
	topoOrder []string,
	metricMap map[string]c.Metric,
	datasets map[string]map[CompanyYearKey]map[string]float64,
	numWorkers int,
) []ScoredRow {
	jobs := make(chan CompanyYearKey, len(allKeys))
	results := make(chan RowResult, len(allKeys))

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for key := range jobs {
				metricsMap := computeScoresForKey(ctx, logger, key, topoOrder, metricMap, datasets)
				results <- RowResult{
					Row: ScoredRow{
						Key:     key,
						Metrics: metricsMap,
					},
					Err: nil,
				}
			}
		}()
	}

	for _, key := range allKeys {
		jobs <- key
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var scoredRows []ScoredRow
	for r := range results {
		if r.Err != nil {
			logger.Sugar().Infow("Error computing scores",
				"company_id", r.Row.Key.CompanyID,
				"year", r.Row.Key.Year,
				"error", r.Err.Error(),
			)
			continue
		}
		scoredRows = append(scoredRows, r.Row)
	}

	sort.Slice(scoredRows, func(i, j int) bool {
		iKey := scoredRows[i].Key
		jKey := scoredRows[j].Key
		if iKey.CompanyID == jKey.CompanyID {
			return iKey.Year < jKey.Year
		}
		return iKey.CompanyID < jKey.CompanyID
	})

	return scoredRows
}

// computeScoresForKey
func computeScoresForKey(
	ctx context.Context,
	logger *zap.Logger,
	key CompanyYearKey,
	topoOrder []string,
	metricMap map[string]c.Metric,
	datasets map[string]map[CompanyYearKey]map[string]float64,
) map[string]float64 {
	metricResults := make(map[string]float64)
	for _, metricName := range topoOrder {
		metricDef := metricMap[metricName]
		val, isNull := evaluateMetric(ctx, logger, metricDef, key, metricResults, datasets)
		if !isNull {
			// store the computed value
			metricResults[metricName] = val
		}
	}
	return metricResults
}

// CalculateScore from file data
func CalculateScore(
	ctx context.Context,
	logger *zap.Logger,
	configFileName string,
	dataService *DataLoaderService,
) (*c.Config, []ScoredRow, error) {

	scoreConfig, err := c.InitScoreConfig(configFileName)
	if err != nil {
		return nil, nil, fmt.Errorf("error initializing score config: %w", err)
	}

	logger.Sugar().Infow("Loaded config",
		"configFileName", configFileName,
		"dataService", dataService,
	)
	// Build the dependency graph & get topological order
	graph, inDegree := buildDependencyGraph(logger, scoreConfig)
	topoOrder, err := topologicalSort(logger, scoreConfig, graph, inDegree)
	if err != nil {
		return nil, nil, fmt.Errorf("failed topological sort: %v", err)
	}
	metricMap := BuildMetricMap(scoreConfig)

	// Load all CSVs (or other files) from "data/" using the injected service
	combined, err := dataService.LoadAllData(ctx, Dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load data from folder: %w", err)
	}

	datasets := make(map[string]map[CompanyYearKey]map[string]float64)
	for logicalName, csvKey := range DatasetKeys {
		data, ok := combined[csvKey]
		if !ok {
			return nil, nil, fmt.Errorf("missing dataset for key=%q", csvKey)
		}
		datasets[logicalName] = data
	}

	allKeys := getAllDataCompanyKeys(datasets)

	scoredResults := parallelComputeScores(ctx, logger, allKeys, topoOrder, metricMap, datasets, NumWorkers)

	logger.Sugar().Infow("Scoring results",
		"results", scoredResults,
		"dataService", dataService,
	)

	return scoreConfig, scoredResults, nil
}

func StreamScores(ctx context.Context,
	logger *zap.Logger,
	allKeys []CompanyYearKey,
	topoOrder []string,
	metricMap map[string]c.Metric,
	datasets map[string]map[CompanyYearKey]map[string]float64,
	numWorkers int) (<-chan ScoredRow, error) {
	out := make(chan ScoredRow)
	go func() {
		defer close(out)
		jobs := make(chan CompanyYearKey, len(allKeys))
		results := make(chan RowResult, len(allKeys))
		var wg sync.WaitGroup
		wg.Add(numWorkers)

		for i := 0; i < numWorkers; i++ {
			go func() {
				defer wg.Done()
				for key := range jobs {
					metricsMap := computeScoresForKey(ctx, logger, key, topoOrder, metricMap, datasets)
					results <- RowResult{
						Row: ScoredRow{
							Key:     key,
							Metrics: metricsMap,
						},
						Err: nil,
					}
				}
			}()
		}

		for _, key := range allKeys {
			jobs <- key
		}
		close(jobs)

		go func() {
			wg.Wait()
			close(results)
		}()

		for r := range results {
			if r.Err != nil {
				logger.Sugar().Errorf("Error computing score for %s %d: %v", r.Row.Key.CompanyID, r.Row.Key.Year, r.Err)
				continue
			}
			select {
			case <-ctx.Done():
				return
			case out <- r.Row:
			}
		}
	}()
	return out, nil
}
