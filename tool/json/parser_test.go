package json

import (
	"testing"
)

///////////////////////////////// Put public apis test cases below ////////////////////////////////

// Test cases for `GetByKeyPath`

var GetByKeyPathTests = []struct {
	desc      string
	paramData []byte
	keyPath   []string
	expect    []byte
	err       error
}{
	{
		desc:      "Expect result is object, and input json is in good format",
		paramData: []byte(`{"a":{"b":{"c":[0,true,false,null,"leonard"]}},"c":"leonard"}`),
		keyPath:   []string{"a", "b"},
		expect:    []byte(`{"c":[0,true,false,null,"leonard"]}`),
		err:       nil,
	}, {
		desc: "Expect result is object, but input json is in arbitrary bad format",
		paramData: []byte(`    {   "a" :   {
                          "b"
                          :
                          {         "c":[0,true,false,null        ,"leonard"]},
                          "c": {
                            "leonard":   123
                          }
                          },"c":    "leonard"    }    `),
		keyPath: []string{"a", "b"},
		expect:  []byte(`{         "c":[0,true,false,null        ,"leonard"]}`),
		err:     nil,
	}, {
		desc: "Expect result is object, but input json is in arbitrary bad format",
		paramData: []byte(`    {   "a" :   {
                          "b"
                          :
                          {         "c":[0,true,false,null        ,"leonard"]},
                          "c": {
                            "leonard":   123
                          }
                          },"c":    "leonard"    }    `),
		keyPath: []string{"a", "c"},
		expect: []byte(`{
                            "leonard":   123
                          }`),
		err: nil,
	}, {
		desc:      "Search by empty string in a good format json, and expect result is simple sequence",
		paramData: []byte(`{"a":{"":0," ":1,"  ":2,"b":{"":{"c":[0,null,"leonard","yeah"]}}}}`),
		keyPath:   []string{"a", ""},
		expect:    []byte(`0`),
		err:       nil,
	}, {
		desc:      "Search by different length empty string in a good format json",
		paramData: []byte(`{"a":{"":0," ":1,"  ":2,"b":{"":{"c":[0,null,"leonard","yeah"]}}}}`),
		keyPath:   []string{"a", " "},
		expect:    []byte(`1`),
		err:       nil,
	}, {
		desc:      "Search by empty string in a good format json, and expect result is an array",
		paramData: []byte(`{"a":{"":0," ":1,"  ":2,"b":{"":{"c":[0,null,"leonard","yeah"]}}}}`),
		keyPath:   []string{"a", "b", "", "c"},
		expect:    []byte(`[0,null,"leonard","yeah"]`),
		err:       nil,
	}, {
		desc: "Search by empty string in arbitrary bad format json, and expect result is an array",
		paramData: []byte(`    {"a"  :{"" : 0
                          ," ":1,"  ":2,
                          "b":{"":{   "c":
                          [0,null,"leonard","yeah"]}
                      }}}`),
		keyPath: []string{"a", "b", "", "c"},
		expect:  []byte(`[0,null,"leonard","yeah"]`),
		err:     nil,
	}, {
		desc: "Expect result is a string",
		paramData: []byte(`    {"a"  :{"" : 0
                          ," ":1,"  ":2,
                          "b":{"":{   "c":
                          [0,null,"leonard","yeah"], "d" : "leonard"   }
                      }}}`),
		keyPath: []string{"a", "b", "", "d"},
		expect:  []byte(`leonard`),
		err:     nil,
	},
}

func TestGetByKeyPath(t *testing.T) {
	for _, test := range GetByKeyPathTests {
		ch := make(chan *V)

		go GetByKeyPath(ch, test.paramData, test.keyPath...)

		// Simultaneous waiting will make it a bit complex to track the correspondence between the result and expectation,
		// so we force to wait one by one!
		//
		// For `GetByKeyPath` will only have one value coming back from the channel.
		res := <-ch

		if res.Err != nil || string(res.V) != string(test.expect) {
			t.Errorf("GetByKeyPath(%s, %s) returned %s, expected %s, and err is %#v", test.paramData, test.keyPath, res.V, test.expect, test.err)
		}
	}
}

// Test cases for `IterateArray`

var IterateArrayTests = []struct {
	desc      string
	paramData []byte
	keyPath   []string
	expect    [][]byte
}{
	{
		desc:      "Iterate over a json style array in a good format json",
		paramData: []byte(`[{"a":[0,true,false,"leonard",null],"b":99,"c":{"d":233}},{"e":666}, "leonard", 666, null]`),
		keyPath:   []string{},
		expect: [][]byte{
			[]byte(`{"a":[0,true,false,"leonard",null],"b":99,"c":{"d":233}}`),
			[]byte(`{"e":666}`),
			[]byte(`leonard`),
			[]byte(`666`),
			[]byte(`null`),
		},
	}, {
		desc: "Iterate over a json style array in a messed up format json, but still valid",
		paramData: []byte(`   [
                        {"a"   :   [0,true,  false  ,"leonard",null]
                        ,"b":99,"c":{"d":233}
                        },   {    "e":666},    "   leonard"
                        , 666, null
                  ]`),
		keyPath: []string{},
		expect: [][]byte{
			[]byte(`{"a"   :   [0,true,  false  ,"leonard",null]
                        ,"b":99,"c":{"d":233}
                        }`),
			[]byte(`{    "e":666}`),
			[]byte(`   leonard`),
			[]byte(`666`),
			[]byte(`null`),
		},
	},
}

func TestIterateArray(t *testing.T) {
	for _, test := range IterateArrayTests {
		ch := make(chan *V)

		go IterateArray(ch, test.paramData, test.keyPath...)

		// Simultaneous waiting will make it a bit complex to track the correspondence between the result and expectation,
		// so we force to wait one by one!
		idx := 0
		for res := range ch {
			if res.Err != nil {
				t.Errorf("IterateArray failed with error %s", res.Err)
				break
			}

			if string(res.V) != string(test.expect[idx]) {
				t.Errorf("IterateArray called back with %s, expected %s", res.V, test.expect[idx])
			}

			idx++
		}
	}
}

// Test cases for `IterateObject`

var IterateObjectTests = []struct {
	desc      string
	paramData []byte
	keyPath   []string
	expectKey [][]byte
	expectVal [][]byte
}{
	{
		desc:      "Iterate over a json object in a good format json",
		paramData: []byte(`{"":{"a":[0,true,"leonard"]},"b":{"c":"miao"},"d":666,"e":"leonard","f":null,"g":[1,true,"leonard"]}`),
		keyPath:   []string{},
		expectKey: [][]byte{
			[]byte(``),
			[]byte(`b`),
			[]byte(`d`),
			[]byte(`e`),
			[]byte(`f`),
			[]byte(`g`),
		},
		expectVal: [][]byte{
			[]byte(`{"a":[0,true,"leonard"]}`),
			[]byte(`{"c":"miao"}`),
			[]byte(`666`),
			[]byte(`leonard`),
			[]byte(`null`),
			[]byte(`[1,true,"leonard"]`),
		},
	}, {
		desc: "Iterate over a json object in a messed up format json, but still valid",
		paramData: []byte(`    {    ""   :{
                                "a"     :[0,
                                true,"leonard"]
                                }   ,"b":{   "c":    "miao"},"d"
                                :   666, "e": "233", "f": 
                                null, " g":  [666, true, "   leonard"]
                        }`),
		keyPath: []string{},
		expectKey: [][]byte{
			[]byte(``),
			[]byte(`b`),
			[]byte(`d`),
			[]byte(`e`),
			[]byte(`f`),
			[]byte(` g`),
		},
		expectVal: [][]byte{
			[]byte(`{
                                "a"     :[0,
                                true,"leonard"]
                                }`),
			[]byte(`{   "c":    "miao"}`),
			[]byte(`666`),
			[]byte(`233`),
			[]byte(`null`),
			[]byte(`[666, true, "   leonard"]`),
		},
	}, {
		desc: "Target function: `findTargetForIterator`",
		paramData: []byte(`{   "a"
                        :{" ":{"c":233
                        ,"d":"leonard","e"    :     null,"f":
                        [null,1,"leonard"],"":666}
                      }}`),
		keyPath: []string{"a", " "},
		expectKey: [][]byte{
			[]byte(`c`),
			[]byte(`d`),
			[]byte(`e`),
			[]byte(`f`),
			[]byte(``),
		},
		expectVal: [][]byte{
			[]byte(`233`),
			[]byte(`leonard`),
			[]byte(`null`),
			[]byte(`[null,1,"leonard"]`),
			[]byte(`666`),
		},
	},
}

func TestIterateObject(t *testing.T) {
	for _, test := range IterateObjectTests {
		ch := make(chan *Kv)

		go IterateObject(ch, test.paramData, test.keyPath...)

		// Simultaneous waiting will make it a bit complex to track the correspondence between the result and expectation,
		// so we force to wait one by one!
		idx := 0
		for res := range ch {
			if res.Err != nil {
				t.Errorf("IterateObject failed with error %s", res.Err)
				break
			}

			if string(res.K) != string(test.expectKey[idx]) {
				t.Errorf("IterateObject called back with key %s, expected key %s", res.K, test.expectKey[idx])
			}
			if string(res.V) != string(test.expectVal[idx]) {
				t.Errorf("IterateObject called back with val %s, expected val %s", res.V, test.expectVal[idx])
			}

			idx++
		}
	}
}

/////////////////////////////// Put internal funcs test cases below ///////////////////////////////

// Test cases for `traverseToStrEnd`

var traverseToStrEndTests = []struct {
	desc      string
	paramData []byte
	expect    int
}{
	{
		desc:      "Strictly rigorous json key",
		paramData: []byte(`high_school"`),
		expect:    12,
	}, {
		desc:      "With some prefixes, white spaces in middle, and suffixes",
		paramData: []byte(`  high_ school  "`),
		expect:    17,
	},
}

func TestTraverseToStrEnd(t *testing.T) {
	for _, test := range traverseToStrEndTests {
		if res := traverseToStrEnd([]byte(test.paramData)); res != test.expect {
			t.Errorf("traverseToStrEnd(%s) returned %d, expected %d", test.paramData, res, test.expect)
		}
	}
}

// Test cases for `traverseToStrEnd`

var traverseToArrOrObjEndTests = []struct {
	desc           string
	paramData      []byte
	paramStartSign byte
	expect         int
}{
	{
		desc:           "A good format json object",
		paramData:      []byte(`{"a":{"b":[1,2,3,null,"leonard",true,false]}}`),
		paramStartSign: '{',
		expect:         45,
	}, {
		desc: "Valid json object, but with arbitrary bad format",
		paramData: []byte(`  {"a":  
                            {"b":[1,2,3,null,"leonard",
                            true,false]  }   }    `),
		paramStartSign: '{',
		expect:         112,
	}, {
		desc:           "Not valid json object",
		paramData:      []byte(`{"a":`),
		paramStartSign: '{',
		expect:         -1,
	}, {
		desc: "Valid json array, but with arbitrary bad format",
		paramData: []byte(`   [   "a"
                              ,1,true,false,"leonard",null,
                              {"b":2,"c":
                              {"d":3}}]       `),
		paramStartSign: '[',
		expect:         152,
	},
}

func TestTraverseToArrOrObjEnd(t *testing.T) {
	for _, test := range traverseToArrOrObjEndTests {
		if res := traverseToArrOrObjEnd(test.paramData, test.paramStartSign); res != test.expect {
			t.Errorf("traverseToArrOrObjEnd(%s, %q) returned %d, expected %d", test.paramData, test.paramStartSign, res, test.expect)
		}
	}
}
