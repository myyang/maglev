package maglev

import "strconv"
import "testing"

func TestMaglev(t *testing.T) {
	ml := NewMaglev([]string{"backend1", "backend2", "backend3"}, 7)

	b, err := ml.Get("Key")
	if err != nil {
		t.Fatalf("Error while Get: %v\n", err)
	}

	bt, _ := ml.Get("Key")

	if bt != b {
		t.Fatalf("Not consistent, %v != %v\n", b, bt)
	}

	err = ml.AddNode("backend4")
	if err != nil {
		t.Fatalf("Error while AddNode: %v\n", err)
	}

	bt2, _ := ml.Get("Key")

	if bt2 != b {
		t.Fatalf("Not consistent after adding, %v != %v\n", b, bt2)
	}

	err = ml.RemoveNode("backend3")
	if err != nil {
		t.Fatalf("Error while RemoveNode: %v\n", err)
	}

	bt3, _ := ml.Get("Key")
	if bt3 == b {
		t.Fatalf("Get key from non-existing node: %v \n", b)
	}
}

func BenchmarkGet(b *testing.B) {
	ml := NewMaglev([]string{"b1", "b2", "b3", "b4", "b5"}, 13)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ml.Get(strconv.Itoa(i))
	}
}
