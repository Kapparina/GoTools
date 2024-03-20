package main

import (
	"encoding/xml"
	"reflect"
	"testing"
)

func TestBuildDataTable(t *testing.T) {
	tests := []struct {
		name string
		rows [][]string
		want DataTable
	}{
		{
			name: "empty rows",
			rows: nil,
			want: DataTable{},
		},
		{
			name: "non-empty row",
			rows: [][]string{
				{"column1", "column2"},
				{"row1Value1", "row1Value2"},
				{"row2Value1", "row2Value2"},
			},
			want: DataTable{
				Rows: []DataRow{
					{
						Columns: []DataColumn{
							{XMLName: xml.Name{Local: "column1"}, Value: "row1Value1"},
							{XMLName: xml.Name{Local: "column2"}, Value: "row1Value2"},
						},
					},
					{
						Columns: []DataColumn{
							{XMLName: xml.Name{Local: "column1"}, Value: "row2Value1"},
							{XMLName: xml.Name{Local: "column2"}, Value: "row2Value2"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildDataTable(&tt.rows); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildDataTable() = %v, want %v", got, tt.want)
			}
		})
	}
}
