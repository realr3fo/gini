package main

import "fmt"

func allCombination(set []string) (subsets [][]string) {
	length := uint(len(set))

	// Go through all possible combinations of objects
	// from 1 (only first object in subset) to 2^length (all objects in subset)
	for subsetBits := 1; subsetBits < (1 << length); subsetBits++ {
		var subset []string

		for object := uint(0); object < length; object++ {
			// checks if object is contained in subset
			// by checking if bit 'object' is set in subsetBits
			if (subsetBits>>object)&1 == 1 {
				// add object to subset
				subset = append(subset, set[object])
			}
		}
		// add subset to subsets
		subsets = append(subsets, subset)
	}
	return subsets
}

func main() {
	setComb := map[int][][]string{}
	combRes := allCombination([]string{"P361","P571","P2250","P1448","P30","P36","P163","P38","P2131","P35"})
	for _, elem := range combRes {
		currLen := len(elem)
		_, ok := setComb[currLen]
		if ok {
			setComb[currLen] = append(setComb[currLen], elem)
		} else {
			setComb[currLen] = [][]string{elem}
		}
	}
	fmt.Println(setComb[1])
	fmt.Println(setComb[2])
	fmt.Println(setComb[3])
	fmt.Println(setComb[4])
	fmt.Println(setComb[5])
	fmt.Println(setComb[6])
	fmt.Println(setComb[7])
	fmt.Println(setComb[8])
	fmt.Println(setComb[9])
	fmt.Println(setComb[10])
}