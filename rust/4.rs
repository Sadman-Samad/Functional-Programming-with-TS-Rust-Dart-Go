

fn main() {
    let plus = |value: u8, func: fn(u8)-> u8| -> u8 {
        func(value)
    };

    let double = |n: u8| -> u8 { n * 2 };

    println!("{}", plus(4, double));
}
