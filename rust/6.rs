static value:[u32;5] = [5,6,6,7,7];


fn main(){

    let v:Vec<u32> = value.iter().filter(|v:&&u32| *v % 2==0).map(|v| v * 2).reduce(|v:u32, n:u32| v*n).into_iter().collect();

    println!("{:?}", v);
}
