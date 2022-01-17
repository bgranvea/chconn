package column_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vahid-sohrabloo/chconn"
	"github.com/vahid-sohrabloo/chconn/column"
)

func TestFixedStringLC(t *testing.T) {
	t.Parallel()

	connString := os.Getenv("CHX_TEST_TCP_CONN_STRING")

	conn, err := chconn.Connect(context.Background(), connString)
	require.NoError(t, err)

	res, err := conn.Exec(context.Background(), `DROP TABLE IF EXISTS test_lc_fixedstring`)
	require.NoError(t, err)
	require.Nil(t, res)

	res, err = conn.Exec(context.Background(), `CREATE TABLE test_lc_fixedstring (
				fixed_lc LowCardinality(FixedString(10)),
				fixed_lc_nullable LowCardinality(Nullable(FixedString(10))),
				fixed_lc_array Array(LowCardinality(FixedString(10))),
				fixed_lc_array_nullable Array(LowCardinality(Nullable(FixedString(10))))
			) Engine=Memory`)

	require.NoError(t, err)
	require.Nil(t, res)

	col := column.NewFixedString(10, false)
	colLC := column.NewLowCardinality(col)

	colNil := column.NewFixedString(10, true)
	colNilLC := column.NewLowCardinality(colNil)

	colArrayValues := column.NewFixedString(10, false)
	collArrayLC := column.NewLowCardinality(colArrayValues)
	colArray := column.NewArray(collArrayLC)

	colArrayValuesNil := column.NewFixedString(10, true)
	collArrayLCNil := column.NewLowCardinality(colArrayValuesNil)
	colArrayNil := column.NewArray(collArrayLCNil)

	var colInsert [][]byte
	var colInsertArray [][][]byte
	var colInsertArrayNil [][]*[]byte
	var colNilInsert []*[]byte

	rows := 10
	for i := 1; i <= rows; i++ {
		val := []byte(fmt.Sprintf("%10d", i))
		valArray := [][]byte{val, []byte(fmt.Sprintf("%10d", i+1))}
		valArrayNil := []*[]byte{&val, nil}

		col.AppendDict(val)
		colInsert = append(colInsert, val)

		// // example insert array
		colInsertArray = append(colInsertArray, valArray)
		colArray.AppendLen(len(valArray))
		for _, v := range valArray {
			colArrayValues.AppendDict(v)
		}

		// example insert nullable array
		colInsertArrayNil = append(colInsertArrayNil, valArrayNil)
		colArrayNil.AppendLen(len(valArrayNil))
		for _, v := range valArrayNil {
			colArrayValuesNil.AppendDictP(v)
		}

		// example add nullable
		if i%2 == 0 {
			colNilInsert = append(colNilInsert, &val)
			if i <= rows/2 {
				// example to add by pointer
				colNil.AppendDictP(&val)
			} else {
				// example to without pointer
				colNil.AppendDict(val)
			}
		} else {
			colNilInsert = append(colNilInsert, nil)
			if i <= rows/2 {
				// example to add by pointer
				colNil.AppendDictP(nil)
			} else {
				// example to add without pointer
				colNil.AppendDictNil()
			}
		}
	}

	err = conn.Insert(context.Background(), `INSERT INTO
		test_lc_fixedstring(fixed_lc,fixed_lc_nullable,fixed_lc_array,fixed_lc_array_nullable)
	VALUES`,
		colLC,
		colNilLC,
		colArray,
		colArrayNil,
	)

	require.NoError(t, err)

	// example read all
	selectStmt, err := conn.Select(context.Background(), `SELECT
		fixed_lc,fixed_lc_nullable,fixed_lc_array,fixed_lc_array_nullable FROM
	test_lc_fixedstring`)
	require.NoError(t, err)
	require.True(t, conn.IsBusy())

	colRead := column.NewFixedString(10, false)
	colLCRead := column.NewLowCardinality(colRead)

	colNilRead := column.NewFixedString(10, true)
	colNilLCRead := column.NewLowCardinality(colNilRead)

	colArrayReadData := column.NewFixedString(10, false)
	colArrayLCRead := column.NewLowCardinality(colArrayReadData)
	colArrayRead := column.NewArray(colArrayLCRead)

	colArrayReadDataNil := column.NewFixedString(10, true)
	colArrayLCReadNil := column.NewLowCardinality(colArrayReadDataNil)
	colArrayReadNil := column.NewArray(colArrayLCReadNil)

	var colDataDict [][]byte
	var colDataKeys []int
	var colData [][]byte

	var colNilDataDict [][]byte
	var colNilDataKeys []int
	var colNilData []*[]byte

	var colArrayDataDict [][]byte
	var colArrayData [][][]byte

	var colArrayDataDictNil [][]byte
	var colArrayDataNil [][]*[]byte

	var colArrayLens []int

	for selectStmt.Next() {
		err = selectStmt.ReadColumns(colLCRead, colNilLCRead, colArrayRead, colArrayReadNil)
		require.NoError(t, err)
		colRead.ReadAll(&colDataDict)
		colLCRead.ReadAll(&colDataKeys)

		for _, k := range colDataKeys {
			colData = append(colData, colDataDict[k])
		}
		colNilRead.ReadAll(&colNilDataDict)
		colNilLCRead.ReadAll(&colNilDataKeys)

		for _, k := range colNilDataKeys {
			// 0 means nil
			if k == 0 {
				colNilData = append(colNilData, nil)
			} else {
				colNilData = append(colNilData, &colNilDataDict[k])
			}
		}

		// read array
		colArrayLens = colArrayLens[:0]
		colArrayRead.ReadAll(&colArrayLens)
		colArrayReadData.ReadAll(&colArrayDataDict)
		for _, l := range colArrayLens {
			arr := make([]int, l)
			arrData := make([][]byte, l)
			colArrayLCRead.Fill(arr)
			for i, k := range arr {
				arrData[i] = colArrayDataDict[k]
			}
			colArrayData = append(colArrayData, arrData)
		}

		// read array nil
		colArrayLens = colArrayLens[:0]
		colArrayReadNil.ReadAll(&colArrayLens)
		colArrayReadDataNil.ReadAll(&colArrayDataDictNil)
		for _, l := range colArrayLens {
			arr := make([]int, l)
			arrData := make([]*[]byte, l)
			colArrayLCReadNil.Fill(arr)
			for i, k := range arr {
				// 0 means nil
				if k == 0 {
					arrData[i] = nil
				} else {
					arrData[i] = &colArrayDataDictNil[k]
				}
			}
			colArrayDataNil = append(colArrayDataNil, arrData)
		}
	}

	require.NoError(t, selectStmt.Err())

	assert.Equal(t, colInsert, colData)
	assert.Equal(t, colNilInsert, colNilData)
	assert.Equal(t, colInsertArray, colArrayData)
	assert.Equal(t, colInsertArrayNil, colArrayDataNil)

	selectStmt.Close()

	// example one by one
	selectStmt, err = conn.Select(context.Background(), `SELECT
		fixed_lc,fixed_lc_nullable
	FROM test_lc_fixedstring`)
	require.NoError(t, err)
	require.True(t, conn.IsBusy())

	colRead = column.NewFixedString(10, false)
	colLCRead = column.NewLowCardinality(colRead)

	colNilRead = column.NewFixedString(10, true)
	colNilLCRead = column.NewLowCardinality(colNilRead)

	colDataDict = colDataDict[:0]
	colData = colData[:0]

	colNilDataDict = colNilDataDict[:0]
	colNilData = colNilData[:0]

	for selectStmt.Next() {
		err = selectStmt.ReadColumns(colLCRead, colNilLCRead)
		require.NoError(t, err)
		colRead.ReadAll(&colDataDict)

		for colLCRead.Next() {
			colData = append(colData, colDataDict[colLCRead.Value()])
		}
		colNilRead.ReadAll(&colNilDataDict)

		for colNilLCRead.Next() {
			k := colNilLCRead.Value()
			// 0 means nil
			if k == 0 {
				colNilData = append(colNilData, nil)
			} else {
				colNilData = append(colNilData, &colNilDataDict[k])
			}
		}
	}

	require.NoError(t, selectStmt.Err())

	selectStmt.Close()

	assert.Equal(t, colInsert, colData)
	assert.Equal(t, colNilInsert, colNilData)
	conn.RawConn().Close()
}
