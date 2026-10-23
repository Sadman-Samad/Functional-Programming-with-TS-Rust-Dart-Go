use std::time::Duration;

use serde::Deserialize;

#[derive(Debug, Deserialize)]
struct Todo {
    #[serde(rename = "userId")]
    _user_id: u32,
    id: u32,
    title: String,
    completed: bool,
}

async fn fetch_todo(client: &reqwest::Client, id: u32) -> Result<Todo, reqwest::Error> {
    client
        .get(format!("https://jsonplaceholder.typicode.com/todos/{id}"))
        .send()
        .await?
        .error_for_status()?
        .json::<Todo>()
        .await
}

async fn background_log(client: reqwest::Client, id: u32) {
    match fetch_todo(&client, id).await {
        Ok(todo) => println!("[spawn] background todo {}: {}", todo.id, todo.title),
        Err(error) => eprintln!("[spawn] background request failed: {error}"),
    }
}

fn heavy_cpu_work(limit: u64) -> u64 {
    (1..=limit).fold(0, |sum, value| sum.wrapping_add(value.wrapping_mul(value)))
}

fn print_todo(label: &str, todo: &Todo) {
    println!(
        "[{label}] {}: {} (completed: {})",
        todo.id, todo.title, todo.completed
    );
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = reqwest::Client::new();

    tokio::spawn(background_log(client.clone(), 9));

    println!("[await] fetching sequentially");
    let first = fetch_todo(&client, 1).await?;
    let second = fetch_todo(&client, 2).await?;
    print_todo("await", &first);
    print_todo("await", &second);

    println!("[join] fetching concurrently");
    let (third, fourth) = tokio::join!(fetch_todo(&client, 3), fetch_todo(&client, 4));
    print_todo("join", &third?);
    print_todo("join", &fourth?);

    println!("[select] racing request against a 2s timeout");
    let winner = tokio::select! {
        result = fetch_todo(&client, 5) => result.map_err(|error| error.to_string()),
        _ = tokio::time::sleep(Duration::from_secs(2)) => Err(String::from("timeout after 2s")),
    };
    match winner {
        Ok(todo) => print_todo("select", &todo),
        Err(message) => println!("[select] {message}"),
    }

    println!("[spawn_blocking] running CPU-bound work off the async threads");
    let checksum = tokio::task::spawn_blocking(|| heavy_cpu_work(10_000_000)).await?;
    println!("[spawn_blocking] checksum = {checksum}");

    tokio::time::sleep(Duration::from_millis(500)).await;
    Ok(())
}
