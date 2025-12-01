package prediction

import (
	"fmt"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

// TimeToMinutes converts "HH:MM" format time to minutes
func TimeToMinutes(timeStr string) (int, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hours: %s", parts[0])
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes: %s", parts[1])
	}

	return hours*60 + minutes, nil
}

// GetProbability calculates visit probability
// Logic equivalent to Python version:
// 1. Cluster data using GMM
// 2. Calculate probability for each cluster
//   - Use cluster center as the mean
//   - Calculate cluster standard deviation
//   - Use normal distribution CDF to calculate probability
//   - Weight probability by (cluster size / weeks)
// 3. Sum probabilities and return
func GetProbability(data []string, time string, weeks int) (float64, error) {
	// Convert time strings to minutes
	dataMinutes := make([]int, 0, len(data))
	for _, d := range data {
		minutes, err := TimeToMinutes(d)
		if err != nil {
			return 0, err
		}
		dataMinutes = append(dataMinutes, minutes)
	}

	timeMinutes, err := TimeToMinutes(time)
	if err != nil {
		return 0, err
	}

	// Special handling for single data point
	if len(dataMinutes) == 1 {
		if timeMinutes >= dataMinutes[0] {
			return 1.0 / float64(weeks), nil
		}
		return 0, nil
	}

	// 1. Cluster data using GMM
	clusters := Clustering(dataMinutes)

	// 2. Calculate probability for each cluster
	var totalProbability float64

	for _, c := range clusters {
		// Single data point in cluster
		if len(c.Data) == 1 {
			if float64(timeMinutes) >= c.Data[0] {
				totalProbability += 1.0 / float64(weeks)
			}
			continue
		}

		// 2-1. Cluster center (mean)
		loc := c.Center

		// 2-2. Calculate cluster standard deviation
		scale := stat.StdDev(c.Data, nil)

		// scale = 0 (all data in cluster are the same)
		if scale == 0 {
			if c.Data[0] == loc && float64(timeMinutes) >= loc {
				totalProbability += 1.0 * (float64(len(c.Data)) / float64(weeks))
			}
			continue
		}

		// 2-3. Calculate probability using normal distribution CDF
		normDist := distuv.Normal{
			Mu:    loc,
			Sigma: scale,
		}
		cdf := normDist.CDF(float64(timeMinutes))

		// 2-4. Weighted probability
		weightedProb := cdf * (float64(len(c.Data)) / float64(weeks))
		totalProbability += weightedProb
	}

	return totalProbability, nil
}

// GetMostLikelyTime finds the most likely time for an activity
// Returns the time (in minutes) with the highest probability density
func GetMostLikelyTime(data []string, weeks int) (int, error) {
	// Convert time strings to minutes
	dataMinutes := make([]int, 0, len(data))
	for _, d := range data {
		minutes, err := TimeToMinutes(d)
		if err != nil {
			return 0, err
		}
		dataMinutes = append(dataMinutes, minutes)
	}

	if len(dataMinutes) == 0 {
		return 0, fmt.Errorf("no data provided")
	}

	if len(dataMinutes) == 1 {
		return dataMinutes[0], nil
	}

	// Cluster data using GMM
	clusters := Clustering(dataMinutes)

	// Find cluster with highest weight
	maxWeight := 0.0
	var bestCluster *ClusteringResult
	for i := range clusters {
		weight := float64(len(clusters[i].Data)) / float64(len(dataMinutes))
		if weight > maxWeight {
			maxWeight = weight
			bestCluster = &clusters[i]
		}
	}

	if bestCluster == nil {
		return 0, fmt.Errorf("clustering failed")
	}

	// Return the center of the most weighted cluster
	return int(bestCluster.Center), nil
}

// MinutesToTime converts minutes to "HH:MM" format
func MinutesToTime(minutes int) string {
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%02d:%02d", hours, mins)
}
