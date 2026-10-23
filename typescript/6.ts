let value: number[] = [3, 6, 6, 7];


function main() {

  let v = value.filter((v) => v % 2 == 0).map((v) => v * 2).reduce((v, n) => n * v);

  console.log(v);
}

main()
