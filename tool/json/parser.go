package json

type V struct {
	V   []byte
	Err error
}

type Kv struct {
	K, V []byte
	Err  error
}

// Search the following structures by using the provided key path in the raw byte slice:
//   1. string
//   2. array / object
//   3. simple sequence, like: 1, true, false, null, undefined etc
// Support key path:
//   1. string
//   2. varying length empty string: "", "  " etc
func GetByKeyPath(ch chan<- *V, data []byte, keys ...string) {
	defer func() {
		recover()
	}()
	defer close(ch)

	if len(data) > 0 && len(keys) > 0 {
		if startIdx, err := searchKeyPath(data, keys...); err != nil {
			ch <- &V{
				Err: err,
			}
		} else {
			switch data[startIdx] {
			case '"':
				if traversedQty := traverseToStrEnd(data[startIdx+1:]); traversedQty == -1 {
					ch <- &V{
						Err: InvalidJson,
					}
				} else {
					ch <- &V{
						V: data[startIdx+1 : startIdx+traversedQty],
					}
				}
			case '[', '{':
				if traversedQty := traverseToArrOrObjEnd(data[startIdx:], data[startIdx]); traversedQty == -1 {
					ch <- &V{
						Err: InvalidJson,
					}
				} else {
					ch <- &V{
						V: data[startIdx : startIdx+traversedQty],
					}
				}
			default:
				if traversedQty := traverseToSimpleSeqEnd(data[startIdx:]); traversedQty == -1 {
					ch <- &V{
						Err: InvalidJson,
					}
				} else {
					ch <- &V{
						V: data[startIdx : startIdx+traversedQty],
					}
				}
			}
		}
	}
}

// Iterate over the array style structure defined by the key path(keys).
// If keys not provide, will try to do so on the `data`.
func IterateArray(ch chan<- *V, data []byte, keys ...string) {
	defer func() {
		recover()
	}()
	defer close(ch)

	if len(data) < 2 {
		ch <- &V{
			Err: InvalidJson,
		}
		return
	}

	target, err := findTargetForIterator(data, '[', keys...)

	if err != nil {
		ch <- &V{
			Err: err,
		}
		return
	}

	idx := 1

	// `target[0]` should be always '['
	// and `target` is guaranteed to be a json array style structure
	for idx < len(target) {

		if traversedQty := traverseToNextVisibleChar(target[idx:]); traversedQty == -1 {
			ch <- &V{
				Err: InvalidJson,
			}
			return
		} else {
			idx += traversedQty
		}

		switch target[idx] {
		case ',':
			idx++
		case ']':
			// handle last ']'
			// consider the case: `   [{},{},{}   ]`
			if idx != len(target)-1 {
				ch <- &V{
					Err: InvalidJson,
				}
				return
			}
			idx++
		case '"':
			if traversedQty := traverseToStrEnd(target[idx+1:]); traversedQty == -1 {
				ch <- &V{
					Err: InvalidJson,
				}
				return
			} else {
				ch <- &V{
					V: target[idx+1 : idx+traversedQty],
				}
				idx += traversedQty + 1
			}
		case '{', '[':
			if traversedQty := traverseToArrOrObjEnd(target[idx:], target[idx]); traversedQty == -1 {
				ch <- &V{
					Err: InvalidJson,
				}
				return
			} else {
				ch <- &V{
					V: target[idx : idx+traversedQty],
				}
				idx += traversedQty
			}
		default:
			if traversedQty := traverseToSimpleSeqEnd(target[idx:]); traversedQty == -1 {
				ch <- &V{
					Err: InvalidJson,
				}
				return
			} else {
				ch <- &V{
					V: target[idx : idx+traversedQty],
				}
				idx += traversedQty
			}
		}

	}
}

// Iterate over the object style structure defined by the key path(keys).
// If keys not provide, will try to do so on the `data`.
func IterateObject(ch chan<- *Kv, data []byte, keys ...string) {
	defer func() {
		recover()
	}()
	defer close(ch)

	if len(data) < 2 {
		ch <- &Kv{
			Err: InvalidJson,
		}
		return
	}

	target, err := findTargetForIterator(data, '{', keys...)

	if err != nil {
		ch <- &Kv{
			Err: err,
		}
		return
	}

	idx, key := 1, ([]byte)(nil)

	// `target[0]` should be always '{'
	// and `target` is guaranteed to be a json object style structure
	for idx < len(target) {

		if traversedQty := traverseToNextVisibleChar(target[idx:]); traversedQty == -1 {
			ch <- &Kv{
				Err: InvalidJson,
			}
			return
		} else {
			idx += traversedQty
		}

		if key == nil {
			// Wait for a key
			switch target[idx] {
			case '"':
				if traversedQty := traverseToStrEnd(target[idx+1:]); traversedQty == -1 {
					ch <- &Kv{
						Err: InvalidJson,
					}
					return
				} else {
					// Only the key which is followed by a valid ':' is a valid structure.
					innerTraversedQty := traverseToNextValidColonForObjKey(target[idx+traversedQty+1:])
					if innerTraversedQty == -1 {
						ch <- &Kv{
							Err: InvalidJson,
						}
						return
					}
					key = target[idx+1 : idx+traversedQty]
					idx += traversedQty + innerTraversedQty + 2
				}
			default:
				ch <- &Kv{
					Err: InvalidJson,
				}
				return
			}
		} else {
			// Wait for a val
			switch target[idx] {
			case '"':
				// It is a `string` value
				if traversedQty := traverseToStrEnd(target[idx+1:]); traversedQty == -1 {
					ch <- &Kv{
						Err: InvalidJson,
					}
					return
				} else {
					// Consume the possibly following ','
					innerTraversedQty := traverseToNextValidCommaForObjVal(target[idx+traversedQty+1:])
					if innerTraversedQty == -1 {
						ch <- &Kv{
							Err: InvalidJson,
						}
						return
					}

					ch <- &Kv{
						K: key,
						V: target[idx+1 : idx+traversedQty],
					}

					idx += traversedQty + innerTraversedQty + 2

					key = nil
				}
			case '{', '[':
				// It is an `object` or `array` value
				if traversedQty := traverseToArrOrObjEnd(target[idx:], target[idx]); traversedQty == -1 {
					ch <- &Kv{
						Err: InvalidJson,
					}
					return
				} else {
					// Consume the possibly following ','
					innerTraversedQty := traverseToNextValidCommaForObjVal(target[idx+traversedQty:])
					if innerTraversedQty == -1 {
						ch <- &Kv{
							Err: InvalidJson,
						}
						return
					}

					ch <- &Kv{
						K: key,
						V: target[idx : idx+traversedQty],
					}

					idx += traversedQty + innerTraversedQty + 1

					key = nil
				}
			default:
				if traversedQty := traverseToSimpleSeqEnd(target[idx:]); traversedQty == -1 {
					ch <- &Kv{
						Err: InvalidJson,
					}
					return
				} else {
					// Consume the possibly following ','
					innerTraversedQty := traverseToNextValidCommaForObjVal(target[idx+traversedQty:])
					if innerTraversedQty == -1 {
						ch <- &Kv{
							Err: InvalidJson,
						}
						return
					}

					ch <- &Kv{
						K: key,
						V: target[idx : idx+traversedQty],
					}

					idx += traversedQty + innerTraversedQty + 1

					key = nil
				}
			}
		}

	}
}

// Add non-exported stuffs below.

func findTargetForIterator(data []byte, targetStartSign byte, keys ...string) (target []byte, err error) {
	if len(keys) > 0 {

		ch := make(chan *V)
		go GetByKeyPath(ch, data, keys...)
		res := <-ch

		if res.Err != nil {
			return nil, res.Err
		}
		if res.V[0] != targetStartSign {
			return nil, InvalidJson
		}
		target = res.V
	} else {
		if traversedQty := traverseToNextVisibleChar(data); traversedQty == -1 {
			return nil, InvalidJson
		} else {
			if data[traversedQty] != targetStartSign {
				return nil, InvalidJson
			}
			if innerTraversedQty := traverseToArrOrObjEnd(data[traversedQty:], targetStartSign); innerTraversedQty == -1 {
				return nil, InvalidJson
			}
			target = data[traversedQty:]
		}
	}

	return target, nil
}

func traverseToNextValidColonForObjKey(data []byte) int {
	for idx, char := range data {
		switch char {
		case ' ', '\n', '\r', '\t':
			continue
		case ':':
			return idx
		default:
			return -1
		}
		idx++
	}
	return -1
}

func traverseToNextValidCommaForObjVal(data []byte) int {
	for idx, char := range data {
		switch char {
		case ' ', '\n', '\r', '\t':
			continue
		case ',', '}':
			return idx
		default:
			return -1
		}
	}
	return -1
}

func traverseToStrEnd(data []byte) int {
	for idx, char := range data {
		if char == '"' {
			return idx + 1
		}
	}
	return -1
}

func traverseToArrOrObjEnd(data []byte, startSign byte) int {
	var endSign byte = '}'
	if startSign == '[' {
		endSign = ']'
	}

	idx, level := 0, 0

	for idx < len(data) {
		switch data[idx] {
		case '"':
			if traversedQty := traverseToStrEnd(data[idx+1:]); traversedQty == -1 {
				return -1
			} else {
				idx += traversedQty
			}
		case startSign:
			level++
		case endSign:
			level--
			if level == 0 {
				return idx + 1
			}
		}
		idx++
	}
	return -1
}

func traverseToNextVisibleChar(data []byte) int {
	for idx, char := range data {
		switch char {
		case ' ', '\n', '\r', '\t':
			continue
		default:
			return idx
		}
	}
	return -1
}

func traverseToSimpleSeqEnd(data []byte) int {
	for idx, char := range data {
		switch char {
		// `..., "a": 233}`
		// `..., 233, true]`
		case ' ', '\n', '\r', '\t', ',', '}', ']':
			return idx
		}
	}
	return -1
}

func searchKeyPath(data []byte, keys ...string) (startIdx int, err error) {
	if len(keys) == 0 {
		return 0, nil
	}

	idx, currLevel, keyLevel := 0, 0, 0

	for idx < len(data) {
		switch data[idx] {
		case '{':
			currLevel++
		case '}':
			currLevel--
		case '[':
			if traversedQty := traverseToArrOrObjEnd(data[idx:], '['); traversedQty == -1 {
				return -1, InvalidJson
			} else {
				idx += (traversedQty - 1)
			}
		case '"':
			idx++
			keyBegin := idx

			if traversedQty := traverseToStrEnd(data[idx:]); traversedQty == -1 {
				return -1, InvalidJson
			} else {
				idx += traversedQty
			}

			keyEnd := idx - 1

			if nextVisibleCharIdx := traverseToNextVisibleChar(data[idx:]); nextVisibleCharIdx == -1 {
				return -1, InvalidJson
			} else {
				idx += nextVisibleCharIdx
			}

			if data[idx] == ':' && keyLevel == currLevel-1 {
				key := data[keyBegin:keyEnd]
				if string(key) == keys[currLevel-1] {
					keyLevel++
					if keyLevel == len(keys) {
						if nextVisibleCharIdx := traverseToNextVisibleChar(data[idx+1:]); nextVisibleCharIdx == -1 {
							return -1, InvalidJson
						} else {
							return idx + 1 + nextVisibleCharIdx, nil
						}
					}
				}
			}
		}
		idx++
	}
	return -1, JsonPathNotFound
}
