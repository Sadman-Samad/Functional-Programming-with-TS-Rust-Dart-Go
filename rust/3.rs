type Op = fn(u32, u32) -> u32;

fn main() {
    let v: Op = |a, b| a + b;

    let can = |a: u32, b: u32, c: Op| -> u32 { c(a, b) };

    println!("{}", can(2, 3, v));
    println!("{}", can(4, 5, |a, b| a * b));
}
