fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_build::compile_protos(option_env!("PUNCHR_PROTO").unwrap_or("../punchr.proto"))?;
    Ok(())
}
