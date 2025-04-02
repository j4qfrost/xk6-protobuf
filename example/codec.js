import protobuf from 'k6/x/protobuf';

const data = open("example.json");
const protoFile = protobuf.load("example/v1/example.proto", "CountryList")

export default function () {

    // Normal encoding decoding
    console.log(protoFile.decode(protoFile.encode(data)))

    // Delimited encoding decoding
    console.log(protoFile.decodeDelimited(protoFile.encodeDelimited(data)))
}

