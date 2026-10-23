let value = 10;
function add(v: number, v2: number) {
  value += v + v2;
  return value;
}
function main() {
  let con = add(20, 40);
  console.log(con);

  if (1000 == 10_00) {
    console.log(true);
  } else false;
}

main();

export {};
