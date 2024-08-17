package main

import (
	"flag"
	"log"
	"math"
	"os"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"google.golang.org/protobuf/proto"
)

var plans = map[int]*v1.License{
	0: {
		Views: 10000,
		Users: 3,
		Sites: 10,
	},
	1: {
		Views: 20000,
		Users: 3,
		Sites: 10,
	},
	3: {
		Views: 50000,
		Users: 3,
		Sites: 10,
	},
	5: {
		Views: 100000,
		Users: 3,
		Sites: 10,
	},
	6: {
		Views: 200000,
		Users: 3,
		Sites: 10,
	},
	7: {
		Views: 500000,
		Users: 3,
		Sites: 10,
	},
	8: {
		Views: 1000000,
		Users: 3,
		Sites: 10,
	},
	9: {
		Views: 10000,
		Users: 10,
		Sites: 50,
	},
	10: {
		Views: 200000,
		Users: 10,
		Sites: 50,
	},
	11: {
		Views: 500000,
		Users: 10,
		Sites: 50,
	},
	12: {
		Views: 1000000,
		Users: 10,
		Sites: 50,
	},
	13: {
		Views: 2000000,
		Users: 10,
		Sites: 50,
	},
	14: {
		Views: 5000000,
		Users: 10,
		Sites: 50,
	},
	16: {
		Views: 10000000,
		Users: 10,
		Sites: 50,
	},
}

var (
	selectPlan = flag.Int("plan", 0, "plan id")
	email      = flag.String("email", "test@example.com", "license email")
	unlimited  = flag.Bool("forever", false, "Forever license")
)

func main() {
	flag.Parse()
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	e, err := openpgp.ReadArmoredKeyRing(f)
	if err != nil {
		log.Fatal(err)
	}
	data, _ := proto.Marshal(getPlan())
	w, err := armor.Encode(os.Stdout, "LICENSE KEY", nil)
	if err != nil {
		log.Fatal(err)
	}
	sw, err := openpgp.Sign(w, e[0], &openpgp.FileHints{IsBinary: true}, nil)
	if err != nil {
		log.Fatal(err)
	}
	sw.Write(data)
	sw.Close()
	w.Close()
}

func getPlan() *v1.License {
	if *unlimited {
		return &v1.License{
			Views:  math.MaxUint64,
			Sites:  math.MaxUint64,
			Users:  math.MaxUint64,
			Expiry: math.MaxUint64,
			Email:  *email,
		}
	}
	year := 364 * 24 * time.Hour
	l := plans[*selectPlan]
	l.Email = *email
	l.Expiry = uint64(time.Now().Add(year).UTC().UnixMilli())
	return l
}
