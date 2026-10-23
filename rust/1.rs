fn addvalue(v: u32, v2: u32) -> u32 {
    let mut value: u32 = 40;
    value += v + v2;
    value
}

fn main() {
    let value = addvalue(20, 40);
    print!("{}", value);
    let value = "hhhh";
    if 1000 == 1_000 {
        println!("{}", value);
    } else {
        println!("None")
    }
}
