import protobuf from 'k6/x/protobuf';

const data = open("example.json");
const protoFile = protobuf.load("example/v1/example.proto", "CountryList")

export default function () {
    //console.log(protoFile.decodeDelimited(protoFile.encodeDelimited(data)))

    //console.log(protoFile.decode(protoFile.encode(data)))

    console.log(protoFile.encodeDelimited(data));

    console.log(protoFile.encode(data))
}

