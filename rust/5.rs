static Value:[u32; 5] = [3,5,6,6,7];


fn main(){
    
    let v:Vec<u32> = Value.iter().map(|v| v*3).collect();

    println!("{:?}", v);


    let flter: Vec<&u32> = Value.iter().filter(|v| **v==6 ).collect();

    println!("{:?}",flter);

    let find = Value.iter().find(|&&c| c==8);

    println!("{:?}",find);

    let reduce = Value.iter().copied().reduce(|v,n| v*n+2);

    println!("{:?}", reduce);
        
}
