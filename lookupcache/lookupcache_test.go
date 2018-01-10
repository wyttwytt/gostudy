package lookupcache

import (
	"fmt"
	"testing"
)

func compareSliceOfSegmentConfig(a, b []SegmentConfig) bool {
	if len(a) == len(b) {
		idx := 0
		for idx < len(a) {
			if a[idx].Id != b[idx].Id {
				return false
			}
			idx++
		}
		return true
	}
	return false
}

var GetSegmentForOrgAndKeyBasicTests = []struct {
	desc     string
	orgKey   string
	paramKey string
	expect   []SegmentConfig
}{
	{
		desc:     "As test case in README.md line 99",
		orgKey:   "6lkb2cv",
		paramKey: "Edu",
		expect:   []SegmentConfig{},
	}, {
		desc:     "As test case in README.md line 103",
		orgKey:   "6lkb2cv",
		paramKey: "sid",
		expect: []SegmentConfig{
			{
				Id: "dem.life.expat",
			},
		},
	},
}

func TestGetSegmentForOrgAndKeyBasic(t *testing.T) {
	for _, test := range GetSegmentForOrgAndKeyBasicTests {
		res := Ec.GetSegmentForOrgAndKey(test.orgKey, test.paramKey)
		if !compareSliceOfSegmentConfig(test.expect, res) {
			t.Errorf("`Ec.GetSegmentForOrgAndKey` failed, returned %s, expected %s", res, test.expect)
		}
	}
}

var GetSegmentForOrgAndKeyAndValBasicTests = []struct {
	desc     string
	orgKey   string
	paramKey string
	paramVal string
	expect   []SegmentConfig
}{
	{
		desc:     "As test case in README.md line 100",
		orgKey:   "6lkb2cv",
		paramKey: "Edu",
		paramVal: "high_school",
		expect: []SegmentConfig{
			{
				Id: "intr.edu.scho",
			}, {
				Id: "intr.edu",
			},
		},
	}, {
		desc:     "As test case in README.md line 101",
		orgKey:   "6lkb2cv",
		paramKey: "Edu",
		paramVal: "bachelors",
		expect: []SegmentConfig{
			{
				Id: "intr.edu",
			},
		},
	}, {
		desc:     "As test case in README.md line 102",
		orgKey:   "6lkb2cv",
		paramKey: "sub",
		paramVal: "Engineering / Architecture",
		expect: []SegmentConfig{
			{
				Id: "dem.emp.con-arch-des",
			}, {
				Id: "dem.emp.eng",
			},
		},
	}, {
		desc:     "As test case in README.md line 104",
		orgKey:   "6lkb2cv",
		paramKey: "sid",
		paramVal: "",
		expect: []SegmentConfig{
			{
				Id: "dem.life.expat",
			},
		},
	}, {
		desc:     "As test case in README.md line 105",
		orgKey:   "6lkb2cv",
		paramKey: "sid",
		paramVal: "anyValueOtherThanAnEmptyString",
		expect:   []SegmentConfig{},
	}, {
		desc:     "As test case in README.md line 106",
		orgKey:   "6lkb2cv",
		paramKey: "gen",
		paramVal: "Female",
		expect: []SegmentConfig{
			{
				Id: "dem.g.f",
			},
		},
	}, {
		desc:     "As test case in README.md line 107",
		orgKey:   "6lkb2cv",
		paramKey: "gen",
		paramVal: "Male",
		expect: []SegmentConfig{
			{
				Id: "dem.g.m",
			},
		},
	}, {
		desc:     "As test case in README.md line 108",
		orgKey:   "6lkb2cv",
		paramKey: "gen",
		paramVal: "anyValueOtherThanMaleFemale",
		expect:   []SegmentConfig{},
	}, {
		desc:     "Extra test case 0",
		orgKey:   "1a9n4ou",
		paramKey: "age",
		paramVal: "18",
		expect: []SegmentConfig{
			{
				Id: "dem.ag.18-24",
			}, {
				Id: "dem.ag.18-20",
			},
		},
	}, {
		desc:     "Extra test case 1",
		orgKey:   "bkie9g1",
		paramKey: "_",
		paramVal: "",
		expect: []SegmentConfig{
			{
				Id: "zz_trash",
			},
		},
	},
}

func TestGetSegmentForOrgAndKeyAndValBasic(t *testing.T) {
	for _, test := range GetSegmentForOrgAndKeyAndValBasicTests {
		res := Ec.GetSegmentForOrgAndKeyAndVal(test.orgKey, test.paramKey, test.paramVal)
		if !compareSliceOfSegmentConfig(test.expect, res) {
			t.Errorf("`Ec.GetSegmentForOrgAndKey` failed, returned %s, expected %s", res, test.expect)
		}
	}
}

func TestEmptyIsAlwaysSame(t *testing.T) {
	res1 := Ec.GetSegmentForOrgAndKeyAndVal("6lkb2cv", "sub", "kids")
	fmt.Println(res1)
	// fmt.Printf("%p, %p", res1, res2)
}
