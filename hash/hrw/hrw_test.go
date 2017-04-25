package hrw

import (
	"encoding/binary"
	"fmt"
	//"hash/crc32"
	"reflect"
	"testing"

	"hash/crc32"

	"time"

	utilsrand "github.com/alex023/basekit/rand"
)

func Example() {
	// given a set of servers
	servers := map[int]string{
		1: "one.example.com",
		2: "two.example.com",
		3: "three.example.com",
		4: "four.example.com",
		5: "five.example.com",
		6: "six.example.com",
	}

	// which can be mapped to integer values
	ids := make([]int, 0, len(servers))
	for id := range servers {
		ids = append(ids, id)
	}

	// HRW can consistently select a uniformly-distributed set of servers for
	// any given key
	key := []byte("/examples/object-key")
	for _, id := range TopN(ids, key, 3) {
		fmt.Printf("trying GET %d %s%s\n", id, servers[id], key)
	}

	// Output:
	// trying GET 1 one.example.com/examples/object-key
	// trying GET 3 three.example.com/examples/object-key
	// trying GET 5 five.example.com/examples/object-key
}

func TestSortByWeight(t *testing.T) {
	key := []byte("hello, world")
	nodes := []int{1, 2, 3, 4, 5}

	actual := SortByWeight(nodes, key)
	expected := []int{5, 4, 2, 1, 3}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Was %#v, but expected %#v", actual, expected)
	}
}

func TestTopN(t *testing.T) {
	key := []byte("hello, world")
	nodes := []int{1, 2, 3, 4, 5}

	actual := TopN(nodes, key, 3)
	expected := []int{5, 4, 2}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Was %#v, but expected %#v", actual, expected)
	}

}

// 测试整形数字在HRW分布的均匀性
func TestUniformDistribution(t *testing.T) {
	//	t.Skip()
	nodes := []int{1, 2, 3, 4, 5}
	counts := make(map[int]int)
	key := make([]byte, 16)
	keys := 1000000

	for i := 0; i < keys; i++ {
		binary.BigEndian.PutUint32(key, uint32(i))
		counts[SortByWeight(nodes, key)[0]]++
	}

	//for node, count := range counts {
	//	t.Logf("Node %d received %10d keys", node, count)
	//}

	mean := float64(keys) / float64(len(nodes))
	delta := mean * 0.02 // 2%
	for node, count := range counts {
		d := mean - float64(count)
		if d > delta || (0-d) > delta {
			t.Errorf(
				"Node %d received %10d keys, expected %v (+/- %v)",
				node, count, mean, delta,
			)
		}
	}
}

// 测试随机字符串在HRW分布的均匀性
func TestUniformDistribution_string(t *testing.T) {
	//	t.Skip()
	nodes := []int{1, 2, 3, 4, 5}
	counts := make(map[int]int)
	key := make([]byte, 16)
	keys := 100000

	for i := 0; i < keys; i++ {
		//随机生成登录名,并通过hash转换为数字,再转换为[]byte
		//一致性hash中,将此部分通过Get方法,做了隐藏
		v := utilsrand.RandomAlphanumeric(30)
		//t.Logf("v=%v \n",[]byte(v))
		v1 := crc32.ChecksumIEEE([]byte(v))
		binary.BigEndian.PutUint32(key, uint32(v1))
		//key=[]byte(v)
		counts[SortByWeight(nodes, key)[0]]++
	}
	//for node, count := range counts {
	//	t.Logf("Node %d received %10d keys", node, count)
	//}
	mean := float64(keys) / float64(len(nodes))
	delta := mean * 0.02 // 2%
	for node, count := range counts {
		d := mean - float64(count)
		if d > delta || (0-d) > delta {
			t.Errorf(
				"Node %d received %12d keys, expected %v (+/- %v)\n",
				node, count, mean, delta,
			)
		}
	}

}

// 测试随机产生的随机字符串,是否每次都能够分配在同一个服务器
//	测试比较耗时，需要约10秒钟时间
func Test_SortUniformDistribution_sameString(t *testing.T) {
	nodes := []int{1, 2, 3, 4, 5}
	key := make([]byte, 16)
	// 产生10000个随机字符串
	for j := 0; j < 10000; j++ {
		// 临时创建,为一个字符串缓存多次hrw的值
		counts := make(map[int]int)
		str := utilsrand.RandomAlphanumeric(20)
		// 每个字符串用HRW计算1万次
		for i := 0; i < 10000; i++ {
			var v uint32
			v = crc32.ChecksumIEEE([]byte(str))

			binary.BigEndian.PutUint32(key, uint32(v))
			counts[SortByWeight(nodes, key)[0]]++
		}
		counter := 0
		for _, x := range counts {
			if x > 0 {
				counter++
			}
		}
		//如果同字符串，被分配到了不同的服务节点，则有问题
		if counter > 1 {
			t.Errorf("当随机字符串=%s,无法确保一致性.分布情况为%v\n", str, counts)
		}
	}

}

// 测试随机产生的随机字符串,是否每次都能够分配在同一个服务器
//	测试比较耗时，需要约10秒钟时间
func Test_SortUniformDistribution_stringOtherTime(t *testing.T) {
	nodes := []int{1, 2, 3, 4, 5}
	key := make([]byte, 16)
	// 产生10个随机字符串
	for j := 0; j < 10; j++ {
		counts := make(map[int]int)
		str := utilsrand.RandomAlphanumeric(20)
		// 每个字符串用HRW计算200次
		for i := 0; i < 200; i++ {
			time.Sleep(time.Millisecond)
			var v uint32
			v = crc32.ChecksumIEEE([]byte(str))

			binary.BigEndian.PutUint32(key, uint32(v))
			counts[SortByWeight(nodes, key)[0]]++
		}
		counter := 0
		for _, x := range counts {
			if x > 0 {
				counter++
			}
		}
		//如果同字符串，被分配到了不同的服务节点，则有问题
		if counter > 1 {
			t.Errorf("当随机字符串=%s,无法确保一致性.分布情况为%v\n", str, counts)
		}
	}

}

func BenchmarkSortByWeight10(b *testing.B) {
	_ = benchmarkSortByWeight(b, 10)
}

func BenchmarkSortByWeight100(b *testing.B) {
	_ = benchmarkSortByWeight(b, 100)
}

func BenchmarkSortByWeight1000(b *testing.B) {
	_ = benchmarkSortByWeight(b, 1000)
}

func benchmarkSortByWeight(b *testing.B, n int) int {
	key := []byte("hello, world")
	servers := make([]int, n)
	for i := 0; i < len(servers); i++ {
		servers[i] = i
	}
	b.ResetTimer()

	var x int
	for i := 0; i < b.N; i++ {
		x += SortByWeight(servers, key)[0]
	}
	return x
}
