struct User {
    name: String,
    roll: f64,
}

async fn m() -> String {
    let f: &str = "kslksl";
    f.to_string()
}

#[tokio::main]
async fn main() {
    let mut v = User {
        name: String::from(""),
        roll: 0.0,
    };

    v.name = String::from("Sadman Samad Lian");
    v.roll = 40.0;

    println!("{} {}", v.name, v.roll);

    println!("{:?}", m().await);
}
