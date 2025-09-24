use solana_program::{
    account_info::{next_account_info, AccountInfo},
    entrypoint,
    entrypoint::ProgramResult,
    msg,
    program_error::ProgramError,
    pubkey::Pubkey,
    sysvar::{rent::Rent, Sysvar},
};

use spl_token::state::{Account as TokenAccount, Mint};

// 程序入口点
entrypoint!(process_instruction);

// 指令类型枚举
#[derive(Debug)]
enum TokenInstruction {
    // 初始化代币铸造
    InitializeMint {
        decimals: u8,
    },
    // 铸造代币
    MintTokens {
        amount: u64,
    },
    // 转移代币
    TransferTokens {
        amount: u64,
    },
}

impl TokenInstruction {
    // 解析指令数据
    fn unpack(input: &[u8]) -> Result<Self, ProgramError> {
        let (&tag, rest) = input.split_first().ok_or(ProgramError::InvalidInstructionData)?;
        
        match tag {
            0 => {
                let decimals = rest[0];
                Ok(Self::InitializeMint { decimals })
            }
            1 => {
                let amount = rest
                    .get(0..8)
                    .and_then(|slice| slice.try_into().ok())
                    .map(u64::from_le_bytes)
                    .ok_or(ProgramError::InvalidInstructionData)?;
                Ok(Self::MintTokens { amount })
            }
            2 => {
                let amount = rest
                    .get(0..8)
                    .and_then(|slice| slice.try_into().ok())
                    .map(u64::from_le_bytes)
                    .ok_or(ProgramError::InvalidInstructionData)?;
                Ok(Self::TransferTokens { amount })
            }
            _ => Err(ProgramError::InvalidInstructionData),
        }
    }
}

// 处理指令的主函数
fn process_instruction(
    program_id: &Pubkey,
    accounts: &[AccountInfo],
    instruction_data: &[u8],
) -> ProgramResult {
    let instruction = TokenInstruction::unpack(instruction_data)?;
    
    match instruction {
        TokenInstruction::InitializeMint { decimals } => {
            msg!("初始化代币铸造");
            initialize_mint(program_id, accounts, decimals)
        }
        TokenInstruction::MintTokens { amount } => {
            msg!("铸造代币: {}", amount);
            mint_tokens(program_id, accounts, amount)
        }
        TokenInstruction::TransferTokens { amount } => {
            msg!("转移代币: {}", amount);
            transfer_tokens(program_id, accounts, amount)
        }
    }
}

// 初始化代币铸造
fn initialize_mint(
    program_id: &Pubkey,
    accounts: &[AccountInfo],
    decimals: u8,
) -> ProgramResult {
    let account_info_iter = &mut accounts.iter();
    
    // 获取所需账户
    let mint_account = next_account_info(account_info_iter)?;
    let mint_authority = next_account_info(account_info_iter)?;
    let rent = &Rent::from_account_info(next_account_info(account_info_iter)?)?;
    
    // 验证账户权限
    if !mint_authority.is_signer {
        return Err(ProgramError::MissingRequiredSignature);
    }
    
    // 初始化铸造账户
    let mint = Mint::new(
        decimals,
        mint_authority.key,
        Some(mint_authority.key),
        0,
    )?;
    
    // 序列化并存储铸造信息
    mint.serialize(&mut *mint_account.data.borrow_mut())?;
    
    // 确保账户有足够的租金豁免
    if !rent.is_exempt(mint_account.lamports(), mint_account.data_len()) {
        return Err(ProgramError::AccountNotRentExempt);
    }
    
    Ok(())
}

// 铸造代币
fn mint_tokens(
    program_id: &Pubkey,
    accounts: &[AccountInfo],
    amount: u64,
) -> ProgramResult {
    let account_info_iter = &mut accounts.iter();
    
    // 获取所需账户
    let mint_account = next_account_info(account_info_iter)?;
    let destination_account = next_account_info(account_info_iter)?;
    let mint_authority = next_account_info(account_info_iter)?;
    
    // 验证账户权限
    if !mint_authority.is_signer {
        return Err(ProgramError::MissingRequiredSignature);
    }
    
    // 加载铸造信息
    let mut mint = Mint::from_account_info(mint_account)?;
    
    // 验证铸造权限
    if *mint_authority.key != mint.mint_authority.ok_or(ProgramError::InvalidMintAuthority)? {
        return Err(ProgramError::InvalidMintAuthority);
    }
    
    // 加载目标账户信息
    let mut destination = TokenAccount::from_account_info(destination_account)?;
    
    // 验证目标账户与铸造账户匹配
    if destination.mint != *mint_account.key {
        return Err(ProgramError::MintMismatch);
    }
    
    // 执行铸造操作
    mint.mint_to(&mut destination, mint_authority, amount)?;
    
    // 保存更新后的状态
    mint.serialize(&mut *mint_account.data.borrow_mut())?;
    destination.serialize(&mut *destination_account.data.borrow_mut())?;
    
    Ok(())
}

// 转移代币
fn transfer_tokens(
    program_id: &Pubkey,
    accounts: &[AccountInfo],
    amount: u64,
) -> ProgramResult {
    let account_info_iter = &mut accounts.iter();
    
    // 获取所需账户
    let source_account = next_account_info(account_info_iter)?;
    let destination_account = next_account_info(account_info_iter)?;
    let owner_account = next_account_info(account_info_iter)?;
    
    // 验证所有者签名
    if !owner_account.is_signer {
        return Err(ProgramError::MissingRequiredSignature);
    }
    
    // 加载源账户和目标账户信息
    let mut source = TokenAccount::from_account_info(source_account)?;
    let mut destination = TokenAccount::from_account_info(destination_account)?;
    
    // 验证两个账户使用相同的铸造
    if source.mint != destination.mint {
        return Err(ProgramError::MintMismatch);
    }
    
    // 验证源账户所有者
    if source.owner != *owner_account.key {
        return Err(ProgramError::InvalidAccountOwner);
    }
    
    // 执行转移操作
    source.transfer(&mut destination, amount)?;
    
    // 保存更新后的状态
    source.serialize(&mut *source_account.data.borrow_mut())?;
    destination.serialize(&mut *destination_account.data.borrow_mut())?;
    
    Ok(())
}
