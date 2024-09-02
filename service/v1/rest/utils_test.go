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

	assert.Equal(t, Btoi(true), 1)
	assert.Equal(t, Btoi(false), 0)
}

func Test_TimeConversionFromDateTimeToUnixAndViceVersa(t *testing.T) {
	/* GIVEN a DateTime object sample
	 * WHEN it is converted to Unix time
	 * AND Unix time is converted to DateTime again
	 * THEN initial sample should be equal with final result
	 */
	t.Parallel()

	initialSample := DateTime{Common{Type: DateTimeStructName}, 2024, 2, 29, 12, 0}

	step, err := dateTimeToUnix(&initialSample)

	assert.NoError(t, err)

	result, err := unixToDateTime(&step)

	assert.NoError(t, err)
	assert.Equal(t, result.Year, initialSample.Year)
	assert.Equal(t, result.Month, initialSample.Month)
	assert.Equal(t, result.Day, initialSample.Day)
	assert.Equal(t, result.Hour, initialSample.Hour)
	assert.Equal(t, result.Minute, initialSample.Minute)
}
