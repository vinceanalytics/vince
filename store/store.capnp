using Go = import "/go.capnp";
@0xda1eebf4d7c0f83e;
$Go.package("store");
$Go.import("store");


struct Calendar{
    visitors @0 :List(Float64);
    visits @1 :List(Float64);
    events @2 :List(Float64);
}