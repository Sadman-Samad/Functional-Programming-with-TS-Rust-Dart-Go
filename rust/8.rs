fn get_value<'a>(value: &'a str) -> &'a str {
    value
}

fn print_value<'a>(value: &'a str) {
    println!("Value: {}", value);
    println!("Memory address: {:p}", value.as_ptr());
}

fn main() {
    let mut name = String::from("Sadman");
    let tname = &name;
    let tname = String::from("Sadman Samad");

    let result = get_value(&name);
    let result2 = get_value(&tname);

    print_value(result);
    print_value(result2);
    print_value(result);
    print_value(result2);
}
