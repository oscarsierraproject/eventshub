package v1rest

// Author: Sebastian Oleksiak (oscarsierraproject@protonmail.com)
// License: The Unlicense
// Created: August 18, 2024

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_BoolToIntConversion(t *testing.T) {
	/* GIVEN a bool value
	 * WHEN it is converted to int value
	 * THEN 'true' should be converted to 1
	 * AND 'false' should be converted to 0
	 */
	t.Parallel()
	var (
		sample_true  bool = true
		sample_false bool = false
	)

	assert.Equal(t, Btoi(sample_true), 1)
	assert.Equal(t, Btoi(sample_false), 0)
}

func Test_TimeConversionFromDateTimeToUnixAndViceVersa(t *testing.T) {
	/* GIVEN a DateTime object sample
	 * WHEN it is converted to Unix time
	 * AND Unix time is converted to DateTime again
	 * THEN initial sample should be equal with final result
	 */
	t.Parallel()
	var initial_sample DateTime = DateTime{Common{Type: DateTimeStructName}, 2024, 2, 29, 12, 0}

	step, _ := dateTimeToUnix(&initial_sample)
	result, _ := unixToDateTime(&step)

	assert.Equal(t, result.Year, initial_sample.Year)
	assert.Equal(t, result.Month, initial_sample.Month)
	assert.Equal(t, result.Day, initial_sample.Day)
	assert.Equal(t, result.Hour, initial_sample.Hour)
	assert.Equal(t, result.Minute, initial_sample.Minute)
}
