package sproto

import (
	"bytes"
	"reflect"
	"testing"
)

type PhoneNumber struct {
	Number *string `sproto:"string,0"`
	Type   *int    `sproto:"integer,1"`
}

type Person struct {
	Name  *string        `sproto:"string,0"`
	Id    *int           `sproto:"integer,1"`
	Email *string        `sproto:"string,2"`
	Phone []*PhoneNumber `sproto:"struct,3,array"`
}

type AddressBook struct {
	Person []*Person `sproto:"struct,0,array"`
}

type Human struct {
	Name     *string  `sproto:"string,0"`
	Age      *int     `sproto:"integer,1"`
	Marital  *bool    `sproto:"boolean,2"`
	Children []*Human `sproto:"struct,3,array"`
}

type Data struct {
	Numbers   []int64   `sproto:"integer,0,array"`
	Bools     []bool    `sproto:"boolean,1,array"`
	Number    *int      `sproto:"integer,2"`
	BigNumber *int64    `sproto:"integer,3"`
	Double    *float64  `sproto:"double,4"`
	Doubles   []float64 `sproto:"double,5,array"`

	Strings []string `sproto:"string,7,array"`
	Bytes   []byte   `sproto:"string,8"`
}

type TestCase struct {
	Name   string
	Struct interface{}
	Data   []byte
}

var abData []byte = []byte{
	1, 0, 0, 0, 122, 0, 0, 0, 68, 0, 0, 0, 4, 0, 0, 0, 34, 78, 1, 0,
	0, 0, 5, 0, 0, 0, 65, 108, 105, 99, 101, 45, 0, 0, 0, 19, 0, 0, 0,
	2, 0, 0, 0, 4, 0, 9, 0, 0, 0, 49, 50, 51, 52, 53, 54, 55, 56, 57,
	18, 0, 0, 0, 2, 0, 0, 0, 6, 0, 8, 0, 0, 0, 56, 55, 54, 53, 52, 51,
	50, 49, 46, 0, 0, 0, 4, 0, 0, 0, 66, 156, 1, 0, 0, 0, 3, 0, 0, 0,
	66, 111, 98, 25, 0, 0, 0, 21, 0, 0, 0, 2, 0, 0, 0, 8, 0, 11, 0, 0,
	0, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48,
}

var abDataPacked []byte = []byte{
	17, 1, 122, 17, 68, 4, 71, 34, 78, 1, 5, 252, 65, 108, 105, 99, 101,
	45, 136, 19, 2, 40, 4, 9, 254, 49, 50, 51, 52, 53, 54, 55, 71, 56, 57,
	18, 2, 20, 6, 8, 255, 0, 56, 55, 54, 53, 52, 51, 50, 49, 17, 46, 4, 71,
	66, 156, 1, 3, 60, 66, 111, 98, 25, 34, 21, 2, 138, 8, 11, 48, 255, 0,
	49, 50, 51, 52, 53, 54, 55, 56, 3, 57, 48,
}

var ab AddressBook = AddressBook{
	Person: []*Person{
		{
			Name: String("Alice"),
			Id:   Int(10000),
			Phone: []*PhoneNumber{
				{
					Number: String("123456789"),
					Type:   Int(1),
				},
				{
					Number: String("87654321"),
					Type:   Int(2),
				},
			},
		},
		{
			Name: String("Bob"),
			Id:   Int(20000),
			Phone: []*PhoneNumber{
				{
					Number: String("01234567890"),
					Type:   Int(3),
				},
			},
		},
	},
}

var testCases []*TestCase = []*TestCase{
	{
		Name: "SimpleStruct",
		Struct: &Human{
			Name:    String("Alice"),
			Age:     Int(13),
			Marital: Bool(false),
		},
		Data: []byte{
			0x03, 0x00, // (fn = 3)
			0x00, 0x00, // (id = 0, value in data part)
			0x1C, 0x00, // (id = 1, value = 13)
			0x02, 0x00, // (id = 2, value = false)
			0x05, 0x00, 0x00, 0x00, // (sizeof "Alice")
			0x41, 0x6C, 0x69, 0x63, 0x65, // ("Alice")
		},
	},
	{
		Name: "StructArray",
		Struct: &Human{
			Name: String("Bob"),
			Age:  Int(40),
			Children: []*Human{
				{
					Name: String("Alice"),
					Age:  Int(13),
				},
				{
					Name: String("Carol"),
					Age:  Int(5),
				},
			},
		},
		Data: []byte{
			0x04, 0x00, // (fn = 4)
			0x00, 0x00, // (id = 0, value in data part)
			0x52, 0x00, // (id = 1, value = 40)
			0x01, 0x00, // (skip id = 2)
			0x00, 0x00, // (id = 3, value in data part)
			0x03, 0x00, 0x00, 0x00, // (sizeof "Bob")
			0x42, 0x6F, 0x62, // ("Bob")
			0x26, 0x00, 0x00, 0x00, // (sizeof children)
			0x0F, 0x00, 0x00, 0x00, // (sizeof child 1)
			0x02, 0x00, //(fn = 2)
			0x00, 0x00, //(id = 0, value in data part)
			0x1C, 0x00, //(id = 1, value = 13)
			0x05, 0x00, 0x00, 0x00, // (sizeof "Alice")
			0x41, 0x6C, 0x69, 0x63, 0x65, //("Alice")
			0x0F, 0x00, 0x00, 0x00, // (sizeof child 2)
			0x02, 0x00, //(fn = 2)
			0x00, 0x00, //(id = 0, value in data part)
			0x0C, 0x00, //(id = 1, value = 5)
			0x05, 0x00, 0x00, 0x00, //(sizeof "Carol")
			0x43, 0x61, 0x72, 0x6F, 0x6C, //("Carol")
		},
	},
	{
		Name: "NumberArray",
		Struct: &Data{
			Numbers: []int64{1, 2, 3, 4, 5},
		},
		Data: []byte{
			0x01, 0x00, // (fn = 1)
			0x00, 0x00, // (id = 0, value in data part)

			0x15, 0x00, 0x00, 0x00, // (sizeof numbers)
			0x04,                   //(sizeof int32)
			0x01, 0x00, 0x00, 0x00, //(1)
			0x02, 0x00, 0x00, 0x00, //(2)
			0x03, 0x00, 0x00, 0x00, //(3)
			0x04, 0x00, 0x00, 0x00, //(4)
			0x05, 0x00, 0x00, 0x00, //(5)
		},
	},
	{
		Name: "BigNumberArray",
		Struct: &Data{
			Numbers: []int64{
				(1 << 32) + 1,
				(1 << 32) + 2,
				(1 << 32) + 3,
			},
		},
		Data: []byte{
			0x01, 0x00, // (fn = 1)
			0x00, 0x00, // (id = 0, value in data part)

			0x19, 0x00, 0x00, 0x00, // (sizeof numbers)
			0x08,                                           //(sizeof int32)
			0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, //((1<<32) + 1)
			0x02, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, //((1<<32) + 2)
			0x03, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, //((1<<32) + 3)
		},
	},
	{
		Name: "BoolArray",
		Struct: &Data{
			Bools: []bool{false, true, false},
		},
		Data: []byte{
			0x02, 0x00, // (fn = 2)
			0x01, 0x00, // (skip id = 0)
			0x00, 0x00, // (id = 1, value in data part)

			0x03, 0x00, 0x00, 0x00, // (sizeof bools)
			0x00, //(false)
			0x01, //(true)
			0x00, //(false)
		},
	},
	{
		Name: "Bytes",
		Struct: &Data{
			Bytes: []byte{0x28, 0x29, 0x30, 0x31},
		},
		Data: []byte{
			0x02, 0x00, // (fn = 2)
			0x0f, 0x00, // (skip id = 7)
			0x00, 0x00, // (id = 8, value in data part)

			0x04, 0x00, 0x00, 0x00, // (sizeof bytes)
			0x28, //(0x28)
			0x29, //(0x29)
			0x30, //(0x30)
			0x31, //(0x31)
		},
	},
	{
		Name: "StringArray",
		Struct: &Data{
			Strings: []string{"Bob", "Alice", "Carol"},
		},
		Data: []byte{
			0x02, 0x00, // (fn = 2)
			0x0d, 0x00, // (skip id = 6)
			0x00, 0x00, // (id = 7, value in data part)

			0x19, 0x00, 0x00, 0x00, // (sizeof []string)
			0x03, 0x00, 0x00, 0x00, // (sizeof "Bob")
			0x42, 0x6F, 0x62, // ("Bob")
			0x05, 0x00, 0x00, 0x00, // (sizeof "Alice")
			0x41, 0x6C, 0x69, 0x63, 0x65, //("Alice")
			0x05, 0x00, 0x00, 0x00, //(sizeof "Carol")
			0x43, 0x61, 0x72, 0x6F, 0x6C, //("Carol")
		},
	},
	{
		Name: "Number",
		Struct: &Data{
			Number:    Int(100000),
			BigNumber: Int64(-10000000000),
		},
		Data: []byte{
			0x03, 0x00, // (fn = 3)
			0x03, 0x00, // (skip id = 1)
			0x00, 0x00, // (id = 2, value in data part)
			0x00, 0x00, // (id = 3, value in data part)

			0x04, 0x00, 0x00, 0x00, //(sizeof number, data part)
			0xA0, 0x86, 0x01, 0x00, //(100000, 32bit integer)

			0x08, 0x00, 0x00, 0x00, //(sizeof bignumber, data part)
			0x00, 0x1C, 0xF4, 0xAB, 0xFD, 0xFF, 0xFF, 0xFF, //(-10000000000, 64bit integer)
		},
	},
	{
		Name: "Double",
		Struct: &Data{
			Double:  Double(0.01171875),
			Doubles: []float64{0.01171875, 23, 4},
		},
		Data: []byte{
			0x03, 0x00, // (fn = 3)
			0x07, 0x00, // (skip id = 3)
			0x00, 0x00, // (id = 4, value in data part)
			0x00, 0x00, // (id = 5, value in data part)

			0x08, 0x00, 0x00, 0x00, // (sizeof number, data part)
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x88, 0x3f, // (0.01171875, 64bit double)

			0x19, 0x00, 0x00, 0x00, // (sizeof doubles)
			0x08,                                           // (sizeof double)
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x88, 0x3f, // (0.01171875, 64bit double)
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x37, 0x40, // (23, 64bit double)
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x40, // (4, 64bit double)
		},
	},
	{
		Name:   "AddressBook",
		Struct: &ab,
		Data:   abData,
	},
}

func TestEncode(t *testing.T) {
	for _, tc := range testCases {
		output, err := Encode(tc.Struct)
		if err != nil {
			t.Fatalf("test case *%s* failed with error:%s", tc.Name, err)
		}
		if !bytes.Equal(output, tc.Data) {
			t.Log("encoded:", output)
			t.Log("expected:", tc.Data)
			t.Fatalf("test case %s failed", tc.Name)
		}
	}
}

func TestDecode(t *testing.T) {
	for _, tc := range testCases {
		sp := reflect.New(reflect.TypeOf(tc.Struct).Elem()).Interface()
		used, err := Decode(tc.Data, sp)
		if err != nil {
			t.Fatalf("test case *%s* failed with error:%s", tc.Name, err)
		}

		if used != len(tc.Data) {
			t.Fatalf("test case *%s* failed: data length mismatch", tc.Name)
		}

		output, err := Encode(sp)
		if err != nil {
			t.Fatalf("test case *%s* failed with error:%s", tc.Name, err)
		}
		if !bytes.Equal(output, tc.Data) {
			t.Log("encoded:", output)
			t.Log("expected:", tc.Data)
			t.Fatalf("test case %s failed", tc.Name)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(&ab)
	}
}

func BenchmarkDecode(b *testing.B) {
	var ab AddressBook
	for i := 0; i < b.N; i++ {
		Decode(abData, &ab)
	}
}

func BenchmarkEncodePacked(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := EncodePacked(&ab)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodePacked(b *testing.B) {
	var ab AddressBook
	for i := 0; i < b.N; i++ {
		if err := DecodePacked(abDataPacked, &ab); err != nil {
			b.Fatal(err)
		}
	}
}
