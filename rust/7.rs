fn con<A, B, C, F, G>(v1: F, v2: G) -> impl Fn(A) -> C
where
    F: Fn(B) -> C,
    G: Fn(A) -> B,
{
    move |value: A| v1(v2(value))
}

fn main() {
    let value = |n: u32| n * 2;
    let value2 = |n: u32| n + 2;

    let v = con(value, value2);

    println!("{}", v(50));
}
