package download

import (
	"fmt"
	"vger/block"
)

type sortFilter struct {
	basicFilter
	from int64
}

func (sf *sortFilter) active() {
	defer sf.closeOutput()

	dbmap := make(map[int64]block.Block)
	nextOutputFrom := sf.from
	for {
		select {
		case b, ok := <-sf.input:
			if !ok {
				return
			}
			// trace(fmt.Sprint("sort filter input:", b.from, b.to))

			dbmap[b.From] = b
			for {
				if b, exist := dbmap[nextOutputFrom]; exist {
					select {
					case sf.output <- b:
						// trace(fmt.Sprint("sort filter write output:", b.from, b.to))

						delete(dbmap, nextOutputFrom)
						nextOutputFrom += int64(len(b.Data))
						break
					case <-sf.quit:
						return
					}
				} else {
					break
				}
			}

		case <-sf.quit:
			fmt.Println("sort output quit")
			return
		}
	}

}
