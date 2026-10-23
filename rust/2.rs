fn main() {
    let value: Vec<u32> = vec![30, 40, 40];

    let mut value2: Vec<u32> = [&value[..], &[6]].concat();

    let value3 = &mut value2;
    value3.push(3);

    println!("{:?}", value2);
}
