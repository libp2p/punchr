pub mod grpc {
    pub struct RegisterRequest {
        #[prost(bytes = "vec", required, tag = "1")]
        pub client_id: ::prost::alloc::vec::Vec<u8>,
        #[prost(string, required, tag = "2")]
        pub agent_version: ::prost::alloc::string::String,
        #[prost(string, repeated, tag = "3")]
        pub protocols: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
        #[prost(string, required, tag = "4")]
        pub api_key: ::prost::alloc::string::String,
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::clone::Clone for RegisterRequest {
        #[inline]
        fn clone(&self) -> RegisterRequest {
            match *self {
                RegisterRequest {
                    client_id: ref __self_0_0,
                    agent_version: ref __self_0_1,
                    protocols: ref __self_0_2,
                    api_key: ref __self_0_3,
                } => RegisterRequest {
                    client_id: ::core::clone::Clone::clone(&(*__self_0_0)),
                    agent_version: ::core::clone::Clone::clone(&(*__self_0_1)),
                    protocols: ::core::clone::Clone::clone(&(*__self_0_2)),
                    api_key: ::core::clone::Clone::clone(&(*__self_0_3)),
                },
            }
        }
    }
    impl ::core::marker::StructuralPartialEq for RegisterRequest {}
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::cmp::PartialEq for RegisterRequest {
        #[inline]
        fn eq(&self, other: &RegisterRequest) -> bool {
            match *other {
                RegisterRequest {
                    client_id: ref __self_1_0,
                    agent_version: ref __self_1_1,
                    protocols: ref __self_1_2,
                    api_key: ref __self_1_3,
                } => match *self {
                    RegisterRequest {
                        client_id: ref __self_0_0,
                        agent_version: ref __self_0_1,
                        protocols: ref __self_0_2,
                        api_key: ref __self_0_3,
                    } => {
                        (*__self_0_0) == (*__self_1_0)
                            && (*__self_0_1) == (*__self_1_1)
                            && (*__self_0_2) == (*__self_1_2)
                            && (*__self_0_3) == (*__self_1_3)
                    }
                },
            }
        }
        #[inline]
        fn ne(&self, other: &RegisterRequest) -> bool {
            match *other {
                RegisterRequest {
                    client_id: ref __self_1_0,
                    agent_version: ref __self_1_1,
                    protocols: ref __self_1_2,
                    api_key: ref __self_1_3,
                } => match *self {
                    RegisterRequest {
                        client_id: ref __self_0_0,
                        agent_version: ref __self_0_1,
                        protocols: ref __self_0_2,
                        api_key: ref __self_0_3,
                    } => {
                        (*__self_0_0) != (*__self_1_0)
                            || (*__self_0_1) != (*__self_1_1)
                            || (*__self_0_2) != (*__self_1_2)
                            || (*__self_0_3) != (*__self_1_3)
                    }
                },
            }
        }
    }
    impl ::prost::Message for RegisterRequest {
        #[allow(unused_variables)]
        fn encode_raw<B>(&self, buf: &mut B)
        where
            B: ::prost::bytes::BufMut,
        {
            ::prost::encoding::bytes::encode(1u32, &self.client_id, buf);
            ::prost::encoding::string::encode(2u32, &self.agent_version, buf);
            ::prost::encoding::string::encode_repeated(3u32, &self.protocols, buf);
            ::prost::encoding::string::encode(4u32, &self.api_key, buf);
        }
        #[allow(unused_variables)]
        fn merge_field<B>(
            &mut self,
            tag: u32,
            wire_type: ::prost::encoding::WireType,
            buf: &mut B,
            ctx: ::prost::encoding::DecodeContext,
        ) -> ::core::result::Result<(), ::prost::DecodeError>
        where
            B: ::prost::bytes::Buf,
        {
            const STRUCT_NAME: &'static str = "RegisterRequest";
            match tag {
                1u32 => {
                    let mut value = &mut self.client_id;
                    ::prost::encoding::bytes::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "client_id");
                            error
                        },
                    )
                }
                2u32 => {
                    let mut value = &mut self.agent_version;
                    ::prost::encoding::string::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "agent_version");
                            error
                        },
                    )
                }
                3u32 => {
                    let mut value = &mut self.protocols;
                    ::prost::encoding::string::merge_repeated(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "protocols");
                            error
                        },
                    )
                }
                4u32 => {
                    let mut value = &mut self.api_key;
                    ::prost::encoding::string::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "api_key");
                            error
                        },
                    )
                }
                _ => ::prost::encoding::skip_field(wire_type, tag, buf, ctx),
            }
        }
        #[inline]
        fn encoded_len(&self) -> usize {
            0 + ::prost::encoding::bytes::encoded_len(1u32, &self.client_id)
                + ::prost::encoding::string::encoded_len(2u32, &self.agent_version)
                + ::prost::encoding::string::encoded_len_repeated(3u32, &self.protocols)
                + ::prost::encoding::string::encoded_len(4u32, &self.api_key)
        }
        fn clear(&mut self) {
            self.client_id.clear();
            self.agent_version.clear();
            self.protocols.clear();
            self.api_key.clear();
        }
    }
    impl ::core::default::Default for RegisterRequest {
        fn default() -> Self {
            RegisterRequest {
                client_id: ::core::default::Default::default(),
                agent_version: ::prost::alloc::string::String::new(),
                protocols: ::prost::alloc::vec::Vec::new(),
                api_key: ::prost::alloc::string::String::new(),
            }
        }
    }
    impl ::core::fmt::Debug for RegisterRequest {
        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
            let mut builder = f.debug_struct("RegisterRequest");
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.client_id)
                };
                builder.field("client_id", &wrapper)
            };
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.agent_version)
                };
                builder.field("agent_version", &wrapper)
            };
            let builder = {
                let wrapper = {
                    struct ScalarWrapper<'a>(
                        &'a ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
                    );
                    impl<'a> ::core::fmt::Debug for ScalarWrapper<'a> {
                        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
                            let mut vec_builder = f.debug_list();
                            for v in self.0 {
                                fn Inner<T>(v: T) -> T {
                                    v
                                }
                                vec_builder.entry(&Inner(v));
                            }
                            vec_builder.finish()
                        }
                    }
                    ScalarWrapper(&self.protocols)
                };
                builder.field("protocols", &wrapper)
            };
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.api_key)
                };
                builder.field("api_key", &wrapper)
            };
            builder.finish()
        }
    }
    pub struct RegisterResponse {
        #[prost(int64, required, tag = "1")]
        pub db_peer_id: i64,
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::clone::Clone for RegisterResponse {
        #[inline]
        fn clone(&self) -> RegisterResponse {
            match *self {
                RegisterResponse {
                    db_peer_id: ref __self_0_0,
                } => RegisterResponse {
                    db_peer_id: ::core::clone::Clone::clone(&(*__self_0_0)),
                },
            }
        }
    }
    impl ::core::marker::StructuralPartialEq for RegisterResponse {}
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::cmp::PartialEq for RegisterResponse {
        #[inline]
        fn eq(&self, other: &RegisterResponse) -> bool {
            match *other {
                RegisterResponse {
                    db_peer_id: ref __self_1_0,
                } => match *self {
                    RegisterResponse {
                        db_peer_id: ref __self_0_0,
                    } => (*__self_0_0) == (*__self_1_0),
                },
            }
        }
        #[inline]
        fn ne(&self, other: &RegisterResponse) -> bool {
            match *other {
                RegisterResponse {
                    db_peer_id: ref __self_1_0,
                } => match *self {
                    RegisterResponse {
                        db_peer_id: ref __self_0_0,
                    } => (*__self_0_0) != (*__self_1_0),
                },
            }
        }
    }
    impl ::prost::Message for RegisterResponse {
        #[allow(unused_variables)]
        fn encode_raw<B>(&self, buf: &mut B)
        where
            B: ::prost::bytes::BufMut,
        {
            ::prost::encoding::int64::encode(1u32, &self.db_peer_id, buf);
        }
        #[allow(unused_variables)]
        fn merge_field<B>(
            &mut self,
            tag: u32,
            wire_type: ::prost::encoding::WireType,
            buf: &mut B,
            ctx: ::prost::encoding::DecodeContext,
        ) -> ::core::result::Result<(), ::prost::DecodeError>
        where
            B: ::prost::bytes::Buf,
        {
            const STRUCT_NAME: &'static str = "RegisterResponse";
            match tag {
                1u32 => {
                    let mut value = &mut self.db_peer_id;
                    ::prost::encoding::int64::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "db_peer_id");
                            error
                        },
                    )
                }
                _ => ::prost::encoding::skip_field(wire_type, tag, buf, ctx),
            }
        }
        #[inline]
        fn encoded_len(&self) -> usize {
            0 + ::prost::encoding::int64::encoded_len(1u32, &self.db_peer_id)
        }
        fn clear(&mut self) {
            self.db_peer_id = 0i64;
        }
    }
    impl ::core::default::Default for RegisterResponse {
        fn default() -> Self {
            RegisterResponse { db_peer_id: 0i64 }
        }
    }
    impl ::core::fmt::Debug for RegisterResponse {
        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
            let mut builder = f.debug_struct("RegisterResponse");
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.db_peer_id)
                };
                builder.field("db_peer_id", &wrapper)
            };
            builder.finish()
        }
    }
    /// There is a challenge when a single punchr client is hole punching a single remote peer via different multi addresses.
    /// Sometimes the NAT mapping stays intact and no new hole punch is attempted. We circumvent this by spawning multiple
    /// libp2p hosts in a single punchr client, each listening on different ports. Hence, when we request a new peer to
    /// hole punch we transmit the host_id for which we request the hole punch + all other libp2p host ids so that we
    /// don't get a peer that is handled by another host already.
    pub struct GetAddrInfoRequest {
        /// Host ID for that the client is requesting a peer to hole punch
        #[prost(bytes = "vec", required, tag = "1")]
        pub host_id: ::prost::alloc::vec::Vec<u8>,
        /// All host IDs that the client is managing
        #[prost(bytes = "vec", repeated, tag = "2")]
        pub all_host_ids: ::prost::alloc::vec::Vec<::prost::alloc::vec::Vec<u8>>,
        /// An authentication key for this request
        #[prost(string, required, tag = "3")]
        pub api_key: ::prost::alloc::string::String,
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::clone::Clone for GetAddrInfoRequest {
        #[inline]
        fn clone(&self) -> GetAddrInfoRequest {
            match *self {
                GetAddrInfoRequest {
                    host_id: ref __self_0_0,
                    all_host_ids: ref __self_0_1,
                    api_key: ref __self_0_2,
                } => GetAddrInfoRequest {
                    host_id: ::core::clone::Clone::clone(&(*__self_0_0)),
                    all_host_ids: ::core::clone::Clone::clone(&(*__self_0_1)),
                    api_key: ::core::clone::Clone::clone(&(*__self_0_2)),
                },
            }
        }
    }
    impl ::core::marker::StructuralPartialEq for GetAddrInfoRequest {}
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::cmp::PartialEq for GetAddrInfoRequest {
        #[inline]
        fn eq(&self, other: &GetAddrInfoRequest) -> bool {
            match *other {
                GetAddrInfoRequest {
                    host_id: ref __self_1_0,
                    all_host_ids: ref __self_1_1,
                    api_key: ref __self_1_2,
                } => match *self {
                    GetAddrInfoRequest {
                        host_id: ref __self_0_0,
                        all_host_ids: ref __self_0_1,
                        api_key: ref __self_0_2,
                    } => {
                        (*__self_0_0) == (*__self_1_0)
                            && (*__self_0_1) == (*__self_1_1)
                            && (*__self_0_2) == (*__self_1_2)
                    }
                },
            }
        }
        #[inline]
        fn ne(&self, other: &GetAddrInfoRequest) -> bool {
            match *other {
                GetAddrInfoRequest {
                    host_id: ref __self_1_0,
                    all_host_ids: ref __self_1_1,
                    api_key: ref __self_1_2,
                } => match *self {
                    GetAddrInfoRequest {
                        host_id: ref __self_0_0,
                        all_host_ids: ref __self_0_1,
                        api_key: ref __self_0_2,
                    } => {
                        (*__self_0_0) != (*__self_1_0)
                            || (*__self_0_1) != (*__self_1_1)
                            || (*__self_0_2) != (*__self_1_2)
                    }
                },
            }
        }
    }
    impl ::prost::Message for GetAddrInfoRequest {
        #[allow(unused_variables)]
        fn encode_raw<B>(&self, buf: &mut B)
        where
            B: ::prost::bytes::BufMut,
        {
            ::prost::encoding::bytes::encode(1u32, &self.host_id, buf);
            ::prost::encoding::bytes::encode_repeated(2u32, &self.all_host_ids, buf);
            ::prost::encoding::string::encode(3u32, &self.api_key, buf);
        }
        #[allow(unused_variables)]
        fn merge_field<B>(
            &mut self,
            tag: u32,
            wire_type: ::prost::encoding::WireType,
            buf: &mut B,
            ctx: ::prost::encoding::DecodeContext,
        ) -> ::core::result::Result<(), ::prost::DecodeError>
        where
            B: ::prost::bytes::Buf,
        {
            const STRUCT_NAME: &'static str = "GetAddrInfoRequest";
            match tag {
                1u32 => {
                    let mut value = &mut self.host_id;
                    ::prost::encoding::bytes::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "host_id");
                            error
                        },
                    )
                }
                2u32 => {
                    let mut value = &mut self.all_host_ids;
                    ::prost::encoding::bytes::merge_repeated(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "all_host_ids");
                            error
                        },
                    )
                }
                3u32 => {
                    let mut value = &mut self.api_key;
                    ::prost::encoding::string::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "api_key");
                            error
                        },
                    )
                }
                _ => ::prost::encoding::skip_field(wire_type, tag, buf, ctx),
            }
        }
        #[inline]
        fn encoded_len(&self) -> usize {
            0 + ::prost::encoding::bytes::encoded_len(1u32, &self.host_id)
                + ::prost::encoding::bytes::encoded_len_repeated(2u32, &self.all_host_ids)
                + ::prost::encoding::string::encoded_len(3u32, &self.api_key)
        }
        fn clear(&mut self) {
            self.host_id.clear();
            self.all_host_ids.clear();
            self.api_key.clear();
        }
    }
    impl ::core::default::Default for GetAddrInfoRequest {
        fn default() -> Self {
            GetAddrInfoRequest {
                host_id: ::core::default::Default::default(),
                all_host_ids: ::prost::alloc::vec::Vec::new(),
                api_key: ::prost::alloc::string::String::new(),
            }
        }
    }
    impl ::core::fmt::Debug for GetAddrInfoRequest {
        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
            let mut builder = f.debug_struct("GetAddrInfoRequest");
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.host_id)
                };
                builder.field("host_id", &wrapper)
            };
            let builder = {
                let wrapper = {
                    struct ScalarWrapper<'a>(
                        &'a ::prost::alloc::vec::Vec<::prost::alloc::vec::Vec<u8>>,
                    );
                    impl<'a> ::core::fmt::Debug for ScalarWrapper<'a> {
                        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
                            let mut vec_builder = f.debug_list();
                            for v in self.0 {
                                fn Inner<T>(v: T) -> T {
                                    v
                                }
                                vec_builder.entry(&Inner(v));
                            }
                            vec_builder.finish()
                        }
                    }
                    ScalarWrapper(&self.all_host_ids)
                };
                builder.field("all_host_ids", &wrapper)
            };
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.api_key)
                };
                builder.field("api_key", &wrapper)
            };
            builder.finish()
        }
    }
    pub struct GetAddrInfoResponse {
        #[prost(bytes = "vec", required, tag = "1")]
        pub remote_id: ::prost::alloc::vec::Vec<u8>,
        #[prost(bytes = "vec", repeated, tag = "2")]
        pub multi_addresses: ::prost::alloc::vec::Vec<::prost::alloc::vec::Vec<u8>>,
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::clone::Clone for GetAddrInfoResponse {
        #[inline]
        fn clone(&self) -> GetAddrInfoResponse {
            match *self {
                GetAddrInfoResponse {
                    remote_id: ref __self_0_0,
                    multi_addresses: ref __self_0_1,
                } => GetAddrInfoResponse {
                    remote_id: ::core::clone::Clone::clone(&(*__self_0_0)),
                    multi_addresses: ::core::clone::Clone::clone(&(*__self_0_1)),
                },
            }
        }
    }
    impl ::core::marker::StructuralPartialEq for GetAddrInfoResponse {}
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::cmp::PartialEq for GetAddrInfoResponse {
        #[inline]
        fn eq(&self, other: &GetAddrInfoResponse) -> bool {
            match *other {
                GetAddrInfoResponse {
                    remote_id: ref __self_1_0,
                    multi_addresses: ref __self_1_1,
                } => match *self {
                    GetAddrInfoResponse {
                        remote_id: ref __self_0_0,
                        multi_addresses: ref __self_0_1,
                    } => (*__self_0_0) == (*__self_1_0) && (*__self_0_1) == (*__self_1_1),
                },
            }
        }
        #[inline]
        fn ne(&self, other: &GetAddrInfoResponse) -> bool {
            match *other {
                GetAddrInfoResponse {
                    remote_id: ref __self_1_0,
                    multi_addresses: ref __self_1_1,
                } => match *self {
                    GetAddrInfoResponse {
                        remote_id: ref __self_0_0,
                        multi_addresses: ref __self_0_1,
                    } => (*__self_0_0) != (*__self_1_0) || (*__self_0_1) != (*__self_1_1),
                },
            }
        }
    }
    impl ::prost::Message for GetAddrInfoResponse {
        #[allow(unused_variables)]
        fn encode_raw<B>(&self, buf: &mut B)
        where
            B: ::prost::bytes::BufMut,
        {
            ::prost::encoding::bytes::encode(1u32, &self.remote_id, buf);
            ::prost::encoding::bytes::encode_repeated(2u32, &self.multi_addresses, buf);
        }
        #[allow(unused_variables)]
        fn merge_field<B>(
            &mut self,
            tag: u32,
            wire_type: ::prost::encoding::WireType,
            buf: &mut B,
            ctx: ::prost::encoding::DecodeContext,
        ) -> ::core::result::Result<(), ::prost::DecodeError>
        where
            B: ::prost::bytes::Buf,
        {
            const STRUCT_NAME: &'static str = "GetAddrInfoResponse";
            match tag {
                1u32 => {
                    let mut value = &mut self.remote_id;
                    ::prost::encoding::bytes::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "remote_id");
                            error
                        },
                    )
                }
                2u32 => {
                    let mut value = &mut self.multi_addresses;
                    ::prost::encoding::bytes::merge_repeated(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "multi_addresses");
                            error
                        },
                    )
                }
                _ => ::prost::encoding::skip_field(wire_type, tag, buf, ctx),
            }
        }
        #[inline]
        fn encoded_len(&self) -> usize {
            0 + ::prost::encoding::bytes::encoded_len(1u32, &self.remote_id)
                + ::prost::encoding::bytes::encoded_len_repeated(2u32, &self.multi_addresses)
        }
        fn clear(&mut self) {
            self.remote_id.clear();
            self.multi_addresses.clear();
        }
    }
    impl ::core::default::Default for GetAddrInfoResponse {
        fn default() -> Self {
            GetAddrInfoResponse {
                remote_id: ::core::default::Default::default(),
                multi_addresses: ::prost::alloc::vec::Vec::new(),
            }
        }
    }
    impl ::core::fmt::Debug for GetAddrInfoResponse {
        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
            let mut builder = f.debug_struct("GetAddrInfoResponse");
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.remote_id)
                };
                builder.field("remote_id", &wrapper)
            };
            let builder = {
                let wrapper = {
                    struct ScalarWrapper<'a>(
                        &'a ::prost::alloc::vec::Vec<::prost::alloc::vec::Vec<u8>>,
                    );
                    impl<'a> ::core::fmt::Debug for ScalarWrapper<'a> {
                        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
                            let mut vec_builder = f.debug_list();
                            for v in self.0 {
                                fn Inner<T>(v: T) -> T {
                                    v
                                }
                                vec_builder.entry(&Inner(v));
                            }
                            vec_builder.finish()
                        }
                    }
                    ScalarWrapper(&self.multi_addresses)
                };
                builder.field("multi_addresses", &wrapper)
            };
            builder.finish()
        }
    }
    pub struct TrackHolePunchRequest {
        /// Peer ID of the requesting punchr client
        #[prost(bytes = "vec", required, tag = "1")]
        pub client_id: ::prost::alloc::vec::Vec<u8>,
        /// Peer ID of the remote peer that was hole punched
        #[prost(bytes = "vec", required, tag = "2")]
        pub remote_id: ::prost::alloc::vec::Vec<u8>,
        /// The multi addresses that were used to attempt a hole punch
        /// (the same that got served in the first place via GetAddrInfo)
        #[prost(bytes = "vec", repeated, tag = "3")]
        pub remote_multi_addresses: ::prost::alloc::vec::Vec<::prost::alloc::vec::Vec<u8>>,
        /// Unix timestamp in nanoseconds of when the connection to the remote peer was initiated
        #[prost(uint64, required, tag = "4")]
        pub connect_started_at: u64,
        /// Unix timestamp in nanoseconds of when the connection to the remote peer via the relay was established (or has failed)
        #[prost(uint64, required, tag = "5")]
        pub connect_ended_at: u64,
        /// Information about each hole punch attempt
        #[prost(message, repeated, tag = "6")]
        pub hole_punch_attempts: ::prost::alloc::vec::Vec<HolePunchAttempt>,
        /// The multi addresses of the open connections AFTER the hole punch process.
        /// This field can be used to track which transport protocols were more successful for hole punching.
        #[prost(bytes = "vec", repeated, tag = "7")]
        pub open_multi_addresses: ::prost::alloc::vec::Vec<::prost::alloc::vec::Vec<u8>>,
        /// Whether the open_multi_addresses contains at least one direct connection
        /// to the remote peer.
        #[prost(bool, required, tag = "8")]
        pub has_direct_conns: bool,
        /// The error that occurred if the hole punch failed
        #[prost(string, optional, tag = "9")]
        pub error: ::core::option::Option<::prost::alloc::string::String>,
        /// The reason why the hole punch ended (direct dial succeeded, protocol error occurred, hole punch procedure finished)
        #[prost(enumeration = "HolePunchOutcome", required, tag = "10")]
        pub outcome: i32,
        /// Unix timestamp in nanoseconds of when the overall hole punch process ended
        #[prost(uint64, required, tag = "11")]
        pub ended_at: u64,
        /// All multi addresses the client is listening on
        #[prost(bytes = "vec", repeated, tag = "12")]
        pub listen_multi_addresses: ::prost::alloc::vec::Vec<::prost::alloc::vec::Vec<u8>>,
        /// An authentication key for this request
        #[prost(string, required, tag = "13")]
        pub api_key: ::prost::alloc::string::String,
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::clone::Clone for TrackHolePunchRequest {
        #[inline]
        fn clone(&self) -> TrackHolePunchRequest {
            match *self {
                TrackHolePunchRequest {
                    client_id: ref __self_0_0,
                    remote_id: ref __self_0_1,
                    remote_multi_addresses: ref __self_0_2,
                    connect_started_at: ref __self_0_3,
                    connect_ended_at: ref __self_0_4,
                    hole_punch_attempts: ref __self_0_5,
                    open_multi_addresses: ref __self_0_6,
                    has_direct_conns: ref __self_0_7,
                    error: ref __self_0_8,
                    outcome: ref __self_0_9,
                    ended_at: ref __self_0_10,
                    listen_multi_addresses: ref __self_0_11,
                    api_key: ref __self_0_12,
                } => TrackHolePunchRequest {
                    client_id: ::core::clone::Clone::clone(&(*__self_0_0)),
                    remote_id: ::core::clone::Clone::clone(&(*__self_0_1)),
                    remote_multi_addresses: ::core::clone::Clone::clone(&(*__self_0_2)),
                    connect_started_at: ::core::clone::Clone::clone(&(*__self_0_3)),
                    connect_ended_at: ::core::clone::Clone::clone(&(*__self_0_4)),
                    hole_punch_attempts: ::core::clone::Clone::clone(&(*__self_0_5)),
                    open_multi_addresses: ::core::clone::Clone::clone(&(*__self_0_6)),
                    has_direct_conns: ::core::clone::Clone::clone(&(*__self_0_7)),
                    error: ::core::clone::Clone::clone(&(*__self_0_8)),
                    outcome: ::core::clone::Clone::clone(&(*__self_0_9)),
                    ended_at: ::core::clone::Clone::clone(&(*__self_0_10)),
                    listen_multi_addresses: ::core::clone::Clone::clone(&(*__self_0_11)),
                    api_key: ::core::clone::Clone::clone(&(*__self_0_12)),
                },
            }
        }
    }
    impl ::core::marker::StructuralPartialEq for TrackHolePunchRequest {}
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::cmp::PartialEq for TrackHolePunchRequest {
        #[inline]
        fn eq(&self, other: &TrackHolePunchRequest) -> bool {
            match *other {
                TrackHolePunchRequest {
                    client_id: ref __self_1_0,
                    remote_id: ref __self_1_1,
                    remote_multi_addresses: ref __self_1_2,
                    connect_started_at: ref __self_1_3,
                    connect_ended_at: ref __self_1_4,
                    hole_punch_attempts: ref __self_1_5,
                    open_multi_addresses: ref __self_1_6,
                    has_direct_conns: ref __self_1_7,
                    error: ref __self_1_8,
                    outcome: ref __self_1_9,
                    ended_at: ref __self_1_10,
                    listen_multi_addresses: ref __self_1_11,
                    api_key: ref __self_1_12,
                } => match *self {
                    TrackHolePunchRequest {
                        client_id: ref __self_0_0,
                        remote_id: ref __self_0_1,
                        remote_multi_addresses: ref __self_0_2,
                        connect_started_at: ref __self_0_3,
                        connect_ended_at: ref __self_0_4,
                        hole_punch_attempts: ref __self_0_5,
                        open_multi_addresses: ref __self_0_6,
                        has_direct_conns: ref __self_0_7,
                        error: ref __self_0_8,
                        outcome: ref __self_0_9,
                        ended_at: ref __self_0_10,
                        listen_multi_addresses: ref __self_0_11,
                        api_key: ref __self_0_12,
                    } => {
                        (*__self_0_0) == (*__self_1_0)
                            && (*__self_0_1) == (*__self_1_1)
                            && (*__self_0_2) == (*__self_1_2)
                            && (*__self_0_3) == (*__self_1_3)
                            && (*__self_0_4) == (*__self_1_4)
                            && (*__self_0_5) == (*__self_1_5)
                            && (*__self_0_6) == (*__self_1_6)
                            && (*__self_0_7) == (*__self_1_7)
                            && (*__self_0_8) == (*__self_1_8)
                            && (*__self_0_9) == (*__self_1_9)
                            && (*__self_0_10) == (*__self_1_10)
                            && (*__self_0_11) == (*__self_1_11)
                            && (*__self_0_12) == (*__self_1_12)
                    }
                },
            }
        }
        #[inline]
        fn ne(&self, other: &TrackHolePunchRequest) -> bool {
            match *other {
                TrackHolePunchRequest {
                    client_id: ref __self_1_0,
                    remote_id: ref __self_1_1,
                    remote_multi_addresses: ref __self_1_2,
                    connect_started_at: ref __self_1_3,
                    connect_ended_at: ref __self_1_4,
                    hole_punch_attempts: ref __self_1_5,
                    open_multi_addresses: ref __self_1_6,
                    has_direct_conns: ref __self_1_7,
                    error: ref __self_1_8,
                    outcome: ref __self_1_9,
                    ended_at: ref __self_1_10,
                    listen_multi_addresses: ref __self_1_11,
                    api_key: ref __self_1_12,
                } => match *self {
                    TrackHolePunchRequest {
                        client_id: ref __self_0_0,
                        remote_id: ref __self_0_1,
                        remote_multi_addresses: ref __self_0_2,
                        connect_started_at: ref __self_0_3,
                        connect_ended_at: ref __self_0_4,
                        hole_punch_attempts: ref __self_0_5,
                        open_multi_addresses: ref __self_0_6,
                        has_direct_conns: ref __self_0_7,
                        error: ref __self_0_8,
                        outcome: ref __self_0_9,
                        ended_at: ref __self_0_10,
                        listen_multi_addresses: ref __self_0_11,
                        api_key: ref __self_0_12,
                    } => {
                        (*__self_0_0) != (*__self_1_0)
                            || (*__self_0_1) != (*__self_1_1)
                            || (*__self_0_2) != (*__self_1_2)
                            || (*__self_0_3) != (*__self_1_3)
                            || (*__self_0_4) != (*__self_1_4)
                            || (*__self_0_5) != (*__self_1_5)
                            || (*__self_0_6) != (*__self_1_6)
                            || (*__self_0_7) != (*__self_1_7)
                            || (*__self_0_8) != (*__self_1_8)
                            || (*__self_0_9) != (*__self_1_9)
                            || (*__self_0_10) != (*__self_1_10)
                            || (*__self_0_11) != (*__self_1_11)
                            || (*__self_0_12) != (*__self_1_12)
                    }
                },
            }
        }
    }
    impl ::prost::Message for TrackHolePunchRequest {
        #[allow(unused_variables)]
        fn encode_raw<B>(&self, buf: &mut B)
        where
            B: ::prost::bytes::BufMut,
        {
            ::prost::encoding::bytes::encode(1u32, &self.client_id, buf);
            ::prost::encoding::bytes::encode(2u32, &self.remote_id, buf);
            ::prost::encoding::bytes::encode_repeated(3u32, &self.remote_multi_addresses, buf);
            ::prost::encoding::uint64::encode(4u32, &self.connect_started_at, buf);
            ::prost::encoding::uint64::encode(5u32, &self.connect_ended_at, buf);
            for msg in &self.hole_punch_attempts {
                ::prost::encoding::message::encode(6u32, msg, buf);
            }
            ::prost::encoding::bytes::encode_repeated(7u32, &self.open_multi_addresses, buf);
            ::prost::encoding::bool::encode(8u32, &self.has_direct_conns, buf);
            if let ::core::option::Option::Some(ref value) = self.error {
                ::prost::encoding::string::encode(9u32, value, buf);
            }
            ::prost::encoding::int32::encode(10u32, &self.outcome, buf);
            ::prost::encoding::uint64::encode(11u32, &self.ended_at, buf);
            ::prost::encoding::bytes::encode_repeated(12u32, &self.listen_multi_addresses, buf);
            ::prost::encoding::string::encode(13u32, &self.api_key, buf);
        }
        #[allow(unused_variables)]
        fn merge_field<B>(
            &mut self,
            tag: u32,
            wire_type: ::prost::encoding::WireType,
            buf: &mut B,
            ctx: ::prost::encoding::DecodeContext,
        ) -> ::core::result::Result<(), ::prost::DecodeError>
        where
            B: ::prost::bytes::Buf,
        {
            const STRUCT_NAME: &'static str = "TrackHolePunchRequest";
            match tag {
                1u32 => {
                    let mut value = &mut self.client_id;
                    ::prost::encoding::bytes::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "client_id");
                            error
                        },
                    )
                }
                2u32 => {
                    let mut value = &mut self.remote_id;
                    ::prost::encoding::bytes::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "remote_id");
                            error
                        },
                    )
                }
                3u32 => {
                    let mut value = &mut self.remote_multi_addresses;
                    ::prost::encoding::bytes::merge_repeated(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "remote_multi_addresses");
                            error
                        },
                    )
                }
                4u32 => {
                    let mut value = &mut self.connect_started_at;
                    ::prost::encoding::uint64::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "connect_started_at");
                            error
                        },
                    )
                }
                5u32 => {
                    let mut value = &mut self.connect_ended_at;
                    ::prost::encoding::uint64::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "connect_ended_at");
                            error
                        },
                    )
                }
                6u32 => {
                    let mut value = &mut self.hole_punch_attempts;
                    ::prost::encoding::message::merge_repeated(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "hole_punch_attempts");
                            error
                        },
                    )
                }
                7u32 => {
                    let mut value = &mut self.open_multi_addresses;
                    ::prost::encoding::bytes::merge_repeated(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "open_multi_addresses");
                            error
                        },
                    )
                }
                8u32 => {
                    let mut value = &mut self.has_direct_conns;
                    ::prost::encoding::bool::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "has_direct_conns");
                            error
                        },
                    )
                }
                9u32 => {
                    let mut value = &mut self.error;
                    ::prost::encoding::string::merge(
                        wire_type,
                        value.get_or_insert_with(::core::default::Default::default),
                        buf,
                        ctx,
                    )
                    .map_err(|mut error| {
                        error.push(STRUCT_NAME, "error");
                        error
                    })
                }
                10u32 => {
                    let mut value = &mut self.outcome;
                    ::prost::encoding::int32::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "outcome");
                            error
                        },
                    )
                }
                11u32 => {
                    let mut value = &mut self.ended_at;
                    ::prost::encoding::uint64::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "ended_at");
                            error
                        },
                    )
                }
                12u32 => {
                    let mut value = &mut self.listen_multi_addresses;
                    ::prost::encoding::bytes::merge_repeated(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "listen_multi_addresses");
                            error
                        },
                    )
                }
                13u32 => {
                    let mut value = &mut self.api_key;
                    ::prost::encoding::string::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "api_key");
                            error
                        },
                    )
                }
                _ => ::prost::encoding::skip_field(wire_type, tag, buf, ctx),
            }
        }
        #[inline]
        fn encoded_len(&self) -> usize {
            0 + ::prost::encoding::bytes::encoded_len(1u32, &self.client_id)
                + ::prost::encoding::bytes::encoded_len(2u32, &self.remote_id)
                + ::prost::encoding::bytes::encoded_len_repeated(3u32, &self.remote_multi_addresses)
                + ::prost::encoding::uint64::encoded_len(4u32, &self.connect_started_at)
                + ::prost::encoding::uint64::encoded_len(5u32, &self.connect_ended_at)
                + ::prost::encoding::message::encoded_len_repeated(6u32, &self.hole_punch_attempts)
                + ::prost::encoding::bytes::encoded_len_repeated(7u32, &self.open_multi_addresses)
                + ::prost::encoding::bool::encoded_len(8u32, &self.has_direct_conns)
                + self.error.as_ref().map_or(0, |value| {
                    ::prost::encoding::string::encoded_len(9u32, value)
                })
                + ::prost::encoding::int32::encoded_len(10u32, &self.outcome)
                + ::prost::encoding::uint64::encoded_len(11u32, &self.ended_at)
                + ::prost::encoding::bytes::encoded_len_repeated(
                    12u32,
                    &self.listen_multi_addresses,
                )
                + ::prost::encoding::string::encoded_len(13u32, &self.api_key)
        }
        fn clear(&mut self) {
            self.client_id.clear();
            self.remote_id.clear();
            self.remote_multi_addresses.clear();
            self.connect_started_at = 0u64;
            self.connect_ended_at = 0u64;
            self.hole_punch_attempts.clear();
            self.open_multi_addresses.clear();
            self.has_direct_conns = false;
            self.error = ::core::option::Option::None;
            self.outcome = HolePunchOutcome::default() as i32;
            self.ended_at = 0u64;
            self.listen_multi_addresses.clear();
            self.api_key.clear();
        }
    }
    impl ::core::default::Default for TrackHolePunchRequest {
        fn default() -> Self {
            TrackHolePunchRequest {
                client_id: ::core::default::Default::default(),
                remote_id: ::core::default::Default::default(),
                remote_multi_addresses: ::prost::alloc::vec::Vec::new(),
                connect_started_at: 0u64,
                connect_ended_at: 0u64,
                hole_punch_attempts: ::core::default::Default::default(),
                open_multi_addresses: ::prost::alloc::vec::Vec::new(),
                has_direct_conns: false,
                error: ::core::option::Option::None,
                outcome: HolePunchOutcome::default() as i32,
                ended_at: 0u64,
                listen_multi_addresses: ::prost::alloc::vec::Vec::new(),
                api_key: ::prost::alloc::string::String::new(),
            }
        }
    }
    impl ::core::fmt::Debug for TrackHolePunchRequest {
        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
            let mut builder = f.debug_struct("TrackHolePunchRequest");
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.client_id)
                };
                builder.field("client_id", &wrapper)
            };
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.remote_id)
                };
                builder.field("remote_id", &wrapper)
            };
            let builder = {
                let wrapper = {
                    struct ScalarWrapper<'a>(
                        &'a ::prost::alloc::vec::Vec<::prost::alloc::vec::Vec<u8>>,
                    );
                    impl<'a> ::core::fmt::Debug for ScalarWrapper<'a> {
                        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
                            let mut vec_builder = f.debug_list();
                            for v in self.0 {
                                fn Inner<T>(v: T) -> T {
                                    v
                                }
                                vec_builder.entry(&Inner(v));
                            }
                            vec_builder.finish()
                        }
                    }
                    ScalarWrapper(&self.remote_multi_addresses)
                };
                builder.field("remote_multi_addresses", &wrapper)
            };
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.connect_started_at)
                };
                builder.field("connect_started_at", &wrapper)
            };
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.connect_ended_at)
                };
                builder.field("connect_ended_at", &wrapper)
            };
            let builder = {
                let wrapper = &self.hole_punch_attempts;
                builder.field("hole_punch_attempts", &wrapper)
            };
            let builder = {
                let wrapper = {
                    struct ScalarWrapper<'a>(
                        &'a ::prost::alloc::vec::Vec<::prost::alloc::vec::Vec<u8>>,
                    );
                    impl<'a> ::core::fmt::Debug for ScalarWrapper<'a> {
                        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
                            let mut vec_builder = f.debug_list();
                            for v in self.0 {
                                fn Inner<T>(v: T) -> T {
                                    v
                                }
                                vec_builder.entry(&Inner(v));
                            }
                            vec_builder.finish()
                        }
                    }
                    ScalarWrapper(&self.open_multi_addresses)
                };
                builder.field("open_multi_addresses", &wrapper)
            };
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.has_direct_conns)
                };
                builder.field("has_direct_conns", &wrapper)
            };
            let builder = {
                let wrapper = {
                    struct ScalarWrapper<'a>(
                        &'a ::core::option::Option<::prost::alloc::string::String>,
                    );
                    impl<'a> ::core::fmt::Debug for ScalarWrapper<'a> {
                        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
                            fn Inner<T>(v: T) -> T {
                                v
                            }
                            ::core::fmt::Debug::fmt(&self.0.as_ref().map(Inner), f)
                        }
                    }
                    ScalarWrapper(&self.error)
                };
                builder.field("error", &wrapper)
            };
            let builder = {
                let wrapper = {
                    struct ScalarWrapper<'a>(&'a i32);
                    impl<'a> ::core::fmt::Debug for ScalarWrapper<'a> {
                        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
                            match HolePunchOutcome::from_i32(*self.0) {
                                None => ::core::fmt::Debug::fmt(&self.0, f),
                                Some(en) => ::core::fmt::Debug::fmt(&en, f),
                            }
                        }
                    }
                    ScalarWrapper(&self.outcome)
                };
                builder.field("outcome", &wrapper)
            };
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.ended_at)
                };
                builder.field("ended_at", &wrapper)
            };
            let builder = {
                let wrapper = {
                    struct ScalarWrapper<'a>(
                        &'a ::prost::alloc::vec::Vec<::prost::alloc::vec::Vec<u8>>,
                    );
                    impl<'a> ::core::fmt::Debug for ScalarWrapper<'a> {
                        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
                            let mut vec_builder = f.debug_list();
                            for v in self.0 {
                                fn Inner<T>(v: T) -> T {
                                    v
                                }
                                vec_builder.entry(&Inner(v));
                            }
                            vec_builder.finish()
                        }
                    }
                    ScalarWrapper(&self.listen_multi_addresses)
                };
                builder.field("listen_multi_addresses", &wrapper)
            };
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.api_key)
                };
                builder.field("api_key", &wrapper)
            };
            builder.finish()
        }
    }
    #[allow(dead_code)]
    impl TrackHolePunchRequest {
        ///Returns the value of `error`, or the default value if `error` is unset.
        pub fn error(&self) -> &str {
            match self.error {
                ::core::option::Option::Some(ref val) => &val[..],
                ::core::option::Option::None => "",
            }
        }
        ///Returns the enum value of `outcome`, or the default if the field is set to an invalid enum value.
        pub fn outcome(&self) -> HolePunchOutcome {
            HolePunchOutcome::from_i32(self.outcome).unwrap_or(HolePunchOutcome::default())
        }
        ///Sets `outcome` to the provided enum value.
        pub fn set_outcome(&mut self, value: HolePunchOutcome) {
            self.outcome = value as i32;
        }
    }
    pub struct TrackHolePunchResponse {}
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::clone::Clone for TrackHolePunchResponse {
        #[inline]
        fn clone(&self) -> TrackHolePunchResponse {
            match *self {
                TrackHolePunchResponse {} => TrackHolePunchResponse {},
            }
        }
    }
    impl ::core::marker::StructuralPartialEq for TrackHolePunchResponse {}
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::cmp::PartialEq for TrackHolePunchResponse {
        #[inline]
        fn eq(&self, other: &TrackHolePunchResponse) -> bool {
            match *other {
                TrackHolePunchResponse {} => match *self {
                    TrackHolePunchResponse {} => true,
                },
            }
        }
    }
    impl ::prost::Message for TrackHolePunchResponse {
        #[allow(unused_variables)]
        fn encode_raw<B>(&self, buf: &mut B)
        where
            B: ::prost::bytes::BufMut,
        {
        }
        #[allow(unused_variables)]
        fn merge_field<B>(
            &mut self,
            tag: u32,
            wire_type: ::prost::encoding::WireType,
            buf: &mut B,
            ctx: ::prost::encoding::DecodeContext,
        ) -> ::core::result::Result<(), ::prost::DecodeError>
        where
            B: ::prost::bytes::Buf,
        {
            match tag {
                _ => ::prost::encoding::skip_field(wire_type, tag, buf, ctx),
            }
        }
        #[inline]
        fn encoded_len(&self) -> usize {
            0
        }
        fn clear(&mut self) {}
    }
    impl ::core::default::Default for TrackHolePunchResponse {
        fn default() -> Self {
            TrackHolePunchResponse {}
        }
    }
    impl ::core::fmt::Debug for TrackHolePunchResponse {
        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
            let mut builder = f.debug_struct("TrackHolePunchResponse");
            builder.finish()
        }
    }
    pub struct HolePunchAttempt {
        /// Unix timestamp in nanoseconds of when the /libp2p/dcutr stream was opened
        #[prost(uint64, required, tag = "1")]
        pub opened_at: u64,
        /// Unix timestamp in nanoseconds of when this hole punching attempt was started
        /// Can be null if hole punch wasn't started, hence `optional`
        #[prost(uint64, optional, tag = "2")]
        pub started_at: ::core::option::Option<u64>,
        /// Unix timestamp in nanoseconds of when this hole punching attempt terminated
        #[prost(uint64, required, tag = "3")]
        pub ended_at: u64,
        /// Start round trip time in seconds that falls out of the `holepunch.StartHolePunchEvt` event
        #[prost(float, optional, tag = "4")]
        pub start_rtt: ::core::option::Option<f32>,
        /// The elapsed time in seconds from start to finish of the hole punch
        #[prost(float, required, tag = "5")]
        pub elapsed_time: f32,
        /// The outcome of the hole punch
        #[prost(enumeration = "HolePunchAttemptOutcome", required, tag = "6")]
        pub outcome: i32,
        /// The error that occurred if the hole punch failed
        #[prost(string, optional, tag = "7")]
        pub error: ::core::option::Option<::prost::alloc::string::String>,
        /// The error that occurred if the connection reversal failed. This is only set of
        /// the multi addresses for the remote peer contained a publicly reachable non-relay multi address
        #[prost(string, optional, tag = "8")]
        pub direct_dial_error: ::core::option::Option<::prost::alloc::string::String>,
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::clone::Clone for HolePunchAttempt {
        #[inline]
        fn clone(&self) -> HolePunchAttempt {
            match *self {
                HolePunchAttempt {
                    opened_at: ref __self_0_0,
                    started_at: ref __self_0_1,
                    ended_at: ref __self_0_2,
                    start_rtt: ref __self_0_3,
                    elapsed_time: ref __self_0_4,
                    outcome: ref __self_0_5,
                    error: ref __self_0_6,
                    direct_dial_error: ref __self_0_7,
                } => HolePunchAttempt {
                    opened_at: ::core::clone::Clone::clone(&(*__self_0_0)),
                    started_at: ::core::clone::Clone::clone(&(*__self_0_1)),
                    ended_at: ::core::clone::Clone::clone(&(*__self_0_2)),
                    start_rtt: ::core::clone::Clone::clone(&(*__self_0_3)),
                    elapsed_time: ::core::clone::Clone::clone(&(*__self_0_4)),
                    outcome: ::core::clone::Clone::clone(&(*__self_0_5)),
                    error: ::core::clone::Clone::clone(&(*__self_0_6)),
                    direct_dial_error: ::core::clone::Clone::clone(&(*__self_0_7)),
                },
            }
        }
    }
    impl ::core::marker::StructuralPartialEq for HolePunchAttempt {}
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::cmp::PartialEq for HolePunchAttempt {
        #[inline]
        fn eq(&self, other: &HolePunchAttempt) -> bool {
            match *other {
                HolePunchAttempt {
                    opened_at: ref __self_1_0,
                    started_at: ref __self_1_1,
                    ended_at: ref __self_1_2,
                    start_rtt: ref __self_1_3,
                    elapsed_time: ref __self_1_4,
                    outcome: ref __self_1_5,
                    error: ref __self_1_6,
                    direct_dial_error: ref __self_1_7,
                } => match *self {
                    HolePunchAttempt {
                        opened_at: ref __self_0_0,
                        started_at: ref __self_0_1,
                        ended_at: ref __self_0_2,
                        start_rtt: ref __self_0_3,
                        elapsed_time: ref __self_0_4,
                        outcome: ref __self_0_5,
                        error: ref __self_0_6,
                        direct_dial_error: ref __self_0_7,
                    } => {
                        (*__self_0_0) == (*__self_1_0)
                            && (*__self_0_1) == (*__self_1_1)
                            && (*__self_0_2) == (*__self_1_2)
                            && (*__self_0_3) == (*__self_1_3)
                            && (*__self_0_4) == (*__self_1_4)
                            && (*__self_0_5) == (*__self_1_5)
                            && (*__self_0_6) == (*__self_1_6)
                            && (*__self_0_7) == (*__self_1_7)
                    }
                },
            }
        }
        #[inline]
        fn ne(&self, other: &HolePunchAttempt) -> bool {
            match *other {
                HolePunchAttempt {
                    opened_at: ref __self_1_0,
                    started_at: ref __self_1_1,
                    ended_at: ref __self_1_2,
                    start_rtt: ref __self_1_3,
                    elapsed_time: ref __self_1_4,
                    outcome: ref __self_1_5,
                    error: ref __self_1_6,
                    direct_dial_error: ref __self_1_7,
                } => match *self {
                    HolePunchAttempt {
                        opened_at: ref __self_0_0,
                        started_at: ref __self_0_1,
                        ended_at: ref __self_0_2,
                        start_rtt: ref __self_0_3,
                        elapsed_time: ref __self_0_4,
                        outcome: ref __self_0_5,
                        error: ref __self_0_6,
                        direct_dial_error: ref __self_0_7,
                    } => {
                        (*__self_0_0) != (*__self_1_0)
                            || (*__self_0_1) != (*__self_1_1)
                            || (*__self_0_2) != (*__self_1_2)
                            || (*__self_0_3) != (*__self_1_3)
                            || (*__self_0_4) != (*__self_1_4)
                            || (*__self_0_5) != (*__self_1_5)
                            || (*__self_0_6) != (*__self_1_6)
                            || (*__self_0_7) != (*__self_1_7)
                    }
                },
            }
        }
    }
    impl ::prost::Message for HolePunchAttempt {
        #[allow(unused_variables)]
        fn encode_raw<B>(&self, buf: &mut B)
        where
            B: ::prost::bytes::BufMut,
        {
            ::prost::encoding::uint64::encode(1u32, &self.opened_at, buf);
            if let ::core::option::Option::Some(ref value) = self.started_at {
                ::prost::encoding::uint64::encode(2u32, value, buf);
            }
            ::prost::encoding::uint64::encode(3u32, &self.ended_at, buf);
            if let ::core::option::Option::Some(ref value) = self.start_rtt {
                ::prost::encoding::float::encode(4u32, value, buf);
            }
            ::prost::encoding::float::encode(5u32, &self.elapsed_time, buf);
            ::prost::encoding::int32::encode(6u32, &self.outcome, buf);
            if let ::core::option::Option::Some(ref value) = self.error {
                ::prost::encoding::string::encode(7u32, value, buf);
            }
            if let ::core::option::Option::Some(ref value) = self.direct_dial_error {
                ::prost::encoding::string::encode(8u32, value, buf);
            }
        }
        #[allow(unused_variables)]
        fn merge_field<B>(
            &mut self,
            tag: u32,
            wire_type: ::prost::encoding::WireType,
            buf: &mut B,
            ctx: ::prost::encoding::DecodeContext,
        ) -> ::core::result::Result<(), ::prost::DecodeError>
        where
            B: ::prost::bytes::Buf,
        {
            const STRUCT_NAME: &'static str = "HolePunchAttempt";
            match tag {
                1u32 => {
                    let mut value = &mut self.opened_at;
                    ::prost::encoding::uint64::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "opened_at");
                            error
                        },
                    )
                }
                2u32 => {
                    let mut value = &mut self.started_at;
                    ::prost::encoding::uint64::merge(
                        wire_type,
                        value.get_or_insert_with(::core::default::Default::default),
                        buf,
                        ctx,
                    )
                    .map_err(|mut error| {
                        error.push(STRUCT_NAME, "started_at");
                        error
                    })
                }
                3u32 => {
                    let mut value = &mut self.ended_at;
                    ::prost::encoding::uint64::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "ended_at");
                            error
                        },
                    )
                }
                4u32 => {
                    let mut value = &mut self.start_rtt;
                    ::prost::encoding::float::merge(
                        wire_type,
                        value.get_or_insert_with(::core::default::Default::default),
                        buf,
                        ctx,
                    )
                    .map_err(|mut error| {
                        error.push(STRUCT_NAME, "start_rtt");
                        error
                    })
                }
                5u32 => {
                    let mut value = &mut self.elapsed_time;
                    ::prost::encoding::float::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "elapsed_time");
                            error
                        },
                    )
                }
                6u32 => {
                    let mut value = &mut self.outcome;
                    ::prost::encoding::int32::merge(wire_type, value, buf, ctx).map_err(
                        |mut error| {
                            error.push(STRUCT_NAME, "outcome");
                            error
                        },
                    )
                }
                7u32 => {
                    let mut value = &mut self.error;
                    ::prost::encoding::string::merge(
                        wire_type,
                        value.get_or_insert_with(::core::default::Default::default),
                        buf,
                        ctx,
                    )
                    .map_err(|mut error| {
                        error.push(STRUCT_NAME, "error");
                        error
                    })
                }
                8u32 => {
                    let mut value = &mut self.direct_dial_error;
                    ::prost::encoding::string::merge(
                        wire_type,
                        value.get_or_insert_with(::core::default::Default::default),
                        buf,
                        ctx,
                    )
                    .map_err(|mut error| {
                        error.push(STRUCT_NAME, "direct_dial_error");
                        error
                    })
                }
                _ => ::prost::encoding::skip_field(wire_type, tag, buf, ctx),
            }
        }
        #[inline]
        fn encoded_len(&self) -> usize {
            0 + ::prost::encoding::uint64::encoded_len(1u32, &self.opened_at)
                + self.started_at.as_ref().map_or(0, |value| {
                    ::prost::encoding::uint64::encoded_len(2u32, value)
                })
                + ::prost::encoding::uint64::encoded_len(3u32, &self.ended_at)
                + self.start_rtt.as_ref().map_or(0, |value| {
                    ::prost::encoding::float::encoded_len(4u32, value)
                })
                + ::prost::encoding::float::encoded_len(5u32, &self.elapsed_time)
                + ::prost::encoding::int32::encoded_len(6u32, &self.outcome)
                + self.error.as_ref().map_or(0, |value| {
                    ::prost::encoding::string::encoded_len(7u32, value)
                })
                + self.direct_dial_error.as_ref().map_or(0, |value| {
                    ::prost::encoding::string::encoded_len(8u32, value)
                })
        }
        fn clear(&mut self) {
            self.opened_at = 0u64;
            self.started_at = ::core::option::Option::None;
            self.ended_at = 0u64;
            self.start_rtt = ::core::option::Option::None;
            self.elapsed_time = 0f32;
            self.outcome = HolePunchAttemptOutcome::default() as i32;
            self.error = ::core::option::Option::None;
            self.direct_dial_error = ::core::option::Option::None;
        }
    }
    impl ::core::default::Default for HolePunchAttempt {
        fn default() -> Self {
            HolePunchAttempt {
                opened_at: 0u64,
                started_at: ::core::option::Option::None,
                ended_at: 0u64,
                start_rtt: ::core::option::Option::None,
                elapsed_time: 0f32,
                outcome: HolePunchAttemptOutcome::default() as i32,
                error: ::core::option::Option::None,
                direct_dial_error: ::core::option::Option::None,
            }
        }
    }
    impl ::core::fmt::Debug for HolePunchAttempt {
        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
            let mut builder = f.debug_struct("HolePunchAttempt");
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.opened_at)
                };
                builder.field("opened_at", &wrapper)
            };
            let builder = {
                let wrapper = {
                    struct ScalarWrapper<'a>(&'a ::core::option::Option<u64>);
                    impl<'a> ::core::fmt::Debug for ScalarWrapper<'a> {
                        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
                            fn Inner<T>(v: T) -> T {
                                v
                            }
                            ::core::fmt::Debug::fmt(&self.0.as_ref().map(Inner), f)
                        }
                    }
                    ScalarWrapper(&self.started_at)
                };
                builder.field("started_at", &wrapper)
            };
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.ended_at)
                };
                builder.field("ended_at", &wrapper)
            };
            let builder = {
                let wrapper = {
                    struct ScalarWrapper<'a>(&'a ::core::option::Option<f32>);
                    impl<'a> ::core::fmt::Debug for ScalarWrapper<'a> {
                        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
                            fn Inner<T>(v: T) -> T {
                                v
                            }
                            ::core::fmt::Debug::fmt(&self.0.as_ref().map(Inner), f)
                        }
                    }
                    ScalarWrapper(&self.start_rtt)
                };
                builder.field("start_rtt", &wrapper)
            };
            let builder = {
                let wrapper = {
                    fn ScalarWrapper<T>(v: T) -> T {
                        v
                    }
                    ScalarWrapper(&self.elapsed_time)
                };
                builder.field("elapsed_time", &wrapper)
            };
            let builder = {
                let wrapper = {
                    struct ScalarWrapper<'a>(&'a i32);
                    impl<'a> ::core::fmt::Debug for ScalarWrapper<'a> {
                        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
                            match HolePunchAttemptOutcome::from_i32(*self.0) {
                                None => ::core::fmt::Debug::fmt(&self.0, f),
                                Some(en) => ::core::fmt::Debug::fmt(&en, f),
                            }
                        }
                    }
                    ScalarWrapper(&self.outcome)
                };
                builder.field("outcome", &wrapper)
            };
            let builder = {
                let wrapper = {
                    struct ScalarWrapper<'a>(
                        &'a ::core::option::Option<::prost::alloc::string::String>,
                    );
                    impl<'a> ::core::fmt::Debug for ScalarWrapper<'a> {
                        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
                            fn Inner<T>(v: T) -> T {
                                v
                            }
                            ::core::fmt::Debug::fmt(&self.0.as_ref().map(Inner), f)
                        }
                    }
                    ScalarWrapper(&self.error)
                };
                builder.field("error", &wrapper)
            };
            let builder = {
                let wrapper = {
                    struct ScalarWrapper<'a>(
                        &'a ::core::option::Option<::prost::alloc::string::String>,
                    );
                    impl<'a> ::core::fmt::Debug for ScalarWrapper<'a> {
                        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
                            fn Inner<T>(v: T) -> T {
                                v
                            }
                            ::core::fmt::Debug::fmt(&self.0.as_ref().map(Inner), f)
                        }
                    }
                    ScalarWrapper(&self.direct_dial_error)
                };
                builder.field("direct_dial_error", &wrapper)
            };
            builder.finish()
        }
    }
    #[allow(dead_code)]
    impl HolePunchAttempt {
        ///Returns the value of `started_at`, or the default value if `started_at` is unset.
        pub fn started_at(&self) -> u64 {
            match self.started_at {
                ::core::option::Option::Some(val) => val,
                ::core::option::Option::None => 0u64,
            }
        }
        ///Returns the value of `start_rtt`, or the default value if `start_rtt` is unset.
        pub fn start_rtt(&self) -> f32 {
            match self.start_rtt {
                ::core::option::Option::Some(val) => val,
                ::core::option::Option::None => 0f32,
            }
        }
        ///Returns the enum value of `outcome`, or the default if the field is set to an invalid enum value.
        pub fn outcome(&self) -> HolePunchAttemptOutcome {
            HolePunchAttemptOutcome::from_i32(self.outcome)
                .unwrap_or(HolePunchAttemptOutcome::default())
        }
        ///Sets `outcome` to the provided enum value.
        pub fn set_outcome(&mut self, value: HolePunchAttemptOutcome) {
            self.outcome = value as i32;
        }
        ///Returns the value of `error`, or the default value if `error` is unset.
        pub fn error(&self) -> &str {
            match self.error {
                ::core::option::Option::Some(ref val) => &val[..],
                ::core::option::Option::None => "",
            }
        }
        ///Returns the value of `direct_dial_error`, or the default value if `direct_dial_error` is unset.
        pub fn direct_dial_error(&self) -> &str {
            match self.direct_dial_error {
                ::core::option::Option::Some(ref val) => &val[..],
                ::core::option::Option::None => "",
            }
        }
    }
    #[repr(i32)]
    pub enum HolePunchOutcome {
        Unknown = 0,
        /// Could not connect to remote peer via relay
        NoConnection = 1,
        /// Hole punch was not initiated by the remote peer
        /// because the /libp2p/dcutr stream was not opened.
        NoStream = 2,
        /// Conditions:
        ///   1. /libp2p/dcutr stream was not opened.
        ///   2. We connected to the remote peer via a relay
        ///   3. We have a direct connection to the remote peer after we have waited for the libp2p/dcutr stream.
        /// Should actually never happen on our side.
        ConnectionReversed = 3,
        /// Hole punch was cancelled by the user
        Cancelled = 4,
        /// The hole punch was attempted several times but failed
        Failed = 5,
        /// The hole punch was performed and successful
        Success = 6,
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::clone::Clone for HolePunchOutcome {
        #[inline]
        fn clone(&self) -> HolePunchOutcome {
            {
                *self
            }
        }
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::marker::Copy for HolePunchOutcome {}
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::fmt::Debug for HolePunchOutcome {
        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
            match (&*self,) {
                (&HolePunchOutcome::Unknown,) => ::core::fmt::Formatter::write_str(f, "Unknown"),
                (&HolePunchOutcome::NoConnection,) => {
                    ::core::fmt::Formatter::write_str(f, "NoConnection")
                }
                (&HolePunchOutcome::NoStream,) => ::core::fmt::Formatter::write_str(f, "NoStream"),
                (&HolePunchOutcome::ConnectionReversed,) => {
                    ::core::fmt::Formatter::write_str(f, "ConnectionReversed")
                }
                (&HolePunchOutcome::Cancelled,) => {
                    ::core::fmt::Formatter::write_str(f, "Cancelled")
                }
                (&HolePunchOutcome::Failed,) => ::core::fmt::Formatter::write_str(f, "Failed"),
                (&HolePunchOutcome::Success,) => ::core::fmt::Formatter::write_str(f, "Success"),
            }
        }
    }
    impl ::core::marker::StructuralPartialEq for HolePunchOutcome {}
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::cmp::PartialEq for HolePunchOutcome {
        #[inline]
        fn eq(&self, other: &HolePunchOutcome) -> bool {
            {
                let __self_vi = ::core::intrinsics::discriminant_value(&*self);
                let __arg_1_vi = ::core::intrinsics::discriminant_value(&*other);
                if true && __self_vi == __arg_1_vi {
                    match (&*self, &*other) {
                        _ => true,
                    }
                } else {
                    false
                }
            }
        }
    }
    impl ::core::marker::StructuralEq for HolePunchOutcome {}
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::cmp::Eq for HolePunchOutcome {
        #[inline]
        #[doc(hidden)]
        #[no_coverage]
        fn assert_receiver_is_total_eq(&self) -> () {
            {}
        }
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::hash::Hash for HolePunchOutcome {
        fn hash<__H: ::core::hash::Hasher>(&self, state: &mut __H) -> () {
            match (&*self,) {
                _ => ::core::hash::Hash::hash(&::core::intrinsics::discriminant_value(self), state),
            }
        }
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::cmp::PartialOrd for HolePunchOutcome {
        #[inline]
        fn partial_cmp(
            &self,
            other: &HolePunchOutcome,
        ) -> ::core::option::Option<::core::cmp::Ordering> {
            {
                let __self_vi = ::core::intrinsics::discriminant_value(&*self);
                let __arg_1_vi = ::core::intrinsics::discriminant_value(&*other);
                if true && __self_vi == __arg_1_vi {
                    match (&*self, &*other) {
                        _ => ::core::option::Option::Some(::core::cmp::Ordering::Equal),
                    }
                } else {
                    ::core::cmp::PartialOrd::partial_cmp(&__self_vi, &__arg_1_vi)
                }
            }
        }
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::cmp::Ord for HolePunchOutcome {
        #[inline]
        fn cmp(&self, other: &HolePunchOutcome) -> ::core::cmp::Ordering {
            {
                let __self_vi = ::core::intrinsics::discriminant_value(&*self);
                let __arg_1_vi = ::core::intrinsics::discriminant_value(&*other);
                if true && __self_vi == __arg_1_vi {
                    match (&*self, &*other) {
                        _ => ::core::cmp::Ordering::Equal,
                    }
                } else {
                    ::core::cmp::Ord::cmp(&__self_vi, &__arg_1_vi)
                }
            }
        }
    }
    impl HolePunchOutcome {
        ///Returns `true` if `value` is a variant of `HolePunchOutcome`.
        pub fn is_valid(value: i32) -> bool {
            match value {
                0 => true,
                1 => true,
                2 => true,
                3 => true,
                4 => true,
                5 => true,
                6 => true,
                _ => false,
            }
        }
        ///Converts an `i32` to a `HolePunchOutcome`, or `None` if `value` is not a valid variant.
        pub fn from_i32(value: i32) -> ::core::option::Option<HolePunchOutcome> {
            match value {
                0 => ::core::option::Option::Some(HolePunchOutcome::Unknown),
                1 => ::core::option::Option::Some(HolePunchOutcome::NoConnection),
                2 => ::core::option::Option::Some(HolePunchOutcome::NoStream),
                3 => ::core::option::Option::Some(HolePunchOutcome::ConnectionReversed),
                4 => ::core::option::Option::Some(HolePunchOutcome::Cancelled),
                5 => ::core::option::Option::Some(HolePunchOutcome::Failed),
                6 => ::core::option::Option::Some(HolePunchOutcome::Success),
                _ => ::core::option::Option::None,
            }
        }
    }
    impl ::core::default::Default for HolePunchOutcome {
        fn default() -> HolePunchOutcome {
            HolePunchOutcome::Unknown
        }
    }
    impl ::core::convert::From<HolePunchOutcome> for i32 {
        fn from(value: HolePunchOutcome) -> i32 {
            value as i32
        }
    }
    #[repr(i32)]
    pub enum HolePunchAttemptOutcome {
        Unknown = 0,
        /// Should never happen on our side. This happens if
        /// the connection reversal from our side succeeded.
        DirectDial = 1,
        /// Can happen if, e.g., the stream was reset mid-flight
        ProtocolError = 2,
        /// The overall hole punch was cancelled by the user
        Cancelled = 3,
        /// The /libp2p/dcutr stream was opened but the hole punch was not initiated in time
        Timeout = 4,
        /// The hole punch was performed but has failed
        Failed = 5,
        /// The hole punch was performed and was successful
        Success = 6,
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::clone::Clone for HolePunchAttemptOutcome {
        #[inline]
        fn clone(&self) -> HolePunchAttemptOutcome {
            {
                *self
            }
        }
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::marker::Copy for HolePunchAttemptOutcome {}
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::fmt::Debug for HolePunchAttemptOutcome {
        fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
            match (&*self,) {
                (&HolePunchAttemptOutcome::Unknown,) => {
                    ::core::fmt::Formatter::write_str(f, "Unknown")
                }
                (&HolePunchAttemptOutcome::DirectDial,) => {
                    ::core::fmt::Formatter::write_str(f, "DirectDial")
                }
                (&HolePunchAttemptOutcome::ProtocolError,) => {
                    ::core::fmt::Formatter::write_str(f, "ProtocolError")
                }
                (&HolePunchAttemptOutcome::Cancelled,) => {
                    ::core::fmt::Formatter::write_str(f, "Cancelled")
                }
                (&HolePunchAttemptOutcome::Timeout,) => {
                    ::core::fmt::Formatter::write_str(f, "Timeout")
                }
                (&HolePunchAttemptOutcome::Failed,) => {
                    ::core::fmt::Formatter::write_str(f, "Failed")
                }
                (&HolePunchAttemptOutcome::Success,) => {
                    ::core::fmt::Formatter::write_str(f, "Success")
                }
            }
        }
    }
    impl ::core::marker::StructuralPartialEq for HolePunchAttemptOutcome {}
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::cmp::PartialEq for HolePunchAttemptOutcome {
        #[inline]
        fn eq(&self, other: &HolePunchAttemptOutcome) -> bool {
            {
                let __self_vi = ::core::intrinsics::discriminant_value(&*self);
                let __arg_1_vi = ::core::intrinsics::discriminant_value(&*other);
                if true && __self_vi == __arg_1_vi {
                    match (&*self, &*other) {
                        _ => true,
                    }
                } else {
                    false
                }
            }
        }
    }
    impl ::core::marker::StructuralEq for HolePunchAttemptOutcome {}
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::cmp::Eq for HolePunchAttemptOutcome {
        #[inline]
        #[doc(hidden)]
        #[no_coverage]
        fn assert_receiver_is_total_eq(&self) -> () {
            {}
        }
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::hash::Hash for HolePunchAttemptOutcome {
        fn hash<__H: ::core::hash::Hasher>(&self, state: &mut __H) -> () {
            match (&*self,) {
                _ => ::core::hash::Hash::hash(&::core::intrinsics::discriminant_value(self), state),
            }
        }
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::cmp::PartialOrd for HolePunchAttemptOutcome {
        #[inline]
        fn partial_cmp(
            &self,
            other: &HolePunchAttemptOutcome,
        ) -> ::core::option::Option<::core::cmp::Ordering> {
            {
                let __self_vi = ::core::intrinsics::discriminant_value(&*self);
                let __arg_1_vi = ::core::intrinsics::discriminant_value(&*other);
                if true && __self_vi == __arg_1_vi {
                    match (&*self, &*other) {
                        _ => ::core::option::Option::Some(::core::cmp::Ordering::Equal),
                    }
                } else {
                    ::core::cmp::PartialOrd::partial_cmp(&__self_vi, &__arg_1_vi)
                }
            }
        }
    }
    #[automatically_derived]
    #[allow(unused_qualifications)]
    impl ::core::cmp::Ord for HolePunchAttemptOutcome {
        #[inline]
        fn cmp(&self, other: &HolePunchAttemptOutcome) -> ::core::cmp::Ordering {
            {
                let __self_vi = ::core::intrinsics::discriminant_value(&*self);
                let __arg_1_vi = ::core::intrinsics::discriminant_value(&*other);
                if true && __self_vi == __arg_1_vi {
                    match (&*self, &*other) {
                        _ => ::core::cmp::Ordering::Equal,
                    }
                } else {
                    ::core::cmp::Ord::cmp(&__self_vi, &__arg_1_vi)
                }
            }
        }
    }
    impl HolePunchAttemptOutcome {
        ///Returns `true` if `value` is a variant of `HolePunchAttemptOutcome`.
        pub fn is_valid(value: i32) -> bool {
            match value {
                0 => true,
                1 => true,
                2 => true,
                3 => true,
                4 => true,
                5 => true,
                6 => true,
                _ => false,
            }
        }
        ///Converts an `i32` to a `HolePunchAttemptOutcome`, or `None` if `value` is not a valid variant.
        pub fn from_i32(value: i32) -> ::core::option::Option<HolePunchAttemptOutcome> {
            match value {
                0 => ::core::option::Option::Some(HolePunchAttemptOutcome::Unknown),
                1 => ::core::option::Option::Some(HolePunchAttemptOutcome::DirectDial),
                2 => ::core::option::Option::Some(HolePunchAttemptOutcome::ProtocolError),
                3 => ::core::option::Option::Some(HolePunchAttemptOutcome::Cancelled),
                4 => ::core::option::Option::Some(HolePunchAttemptOutcome::Timeout),
                5 => ::core::option::Option::Some(HolePunchAttemptOutcome::Failed),
                6 => ::core::option::Option::Some(HolePunchAttemptOutcome::Success),
                _ => ::core::option::Option::None,
            }
        }
    }
    impl ::core::default::Default for HolePunchAttemptOutcome {
        fn default() -> HolePunchAttemptOutcome {
            HolePunchAttemptOutcome::Unknown
        }
    }
    impl ::core::convert::From<HolePunchAttemptOutcome> for i32 {
        fn from(value: HolePunchAttemptOutcome) -> i32 {
            value as i32
        }
    }
    /// Generated client implementations.
    pub mod punchr_service_client {
        #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
        use tonic::codegen::*;
        pub struct PunchrServiceClient<T> {
            inner: tonic::client::Grpc<T>,
        }
        #[automatically_derived]
        #[allow(unused_qualifications)]
        impl<T: ::core::fmt::Debug> ::core::fmt::Debug for PunchrServiceClient<T> {
            fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
                match *self {
                    PunchrServiceClient {
                        inner: ref __self_0_0,
                    } => {
                        let debug_trait_builder =
                            &mut ::core::fmt::Formatter::debug_struct(f, "PunchrServiceClient");
                        let _ = ::core::fmt::DebugStruct::field(
                            debug_trait_builder,
                            "inner",
                            &&(*__self_0_0),
                        );
                        ::core::fmt::DebugStruct::finish(debug_trait_builder)
                    }
                }
            }
        }
        #[automatically_derived]
        #[allow(unused_qualifications)]
        impl<T: ::core::clone::Clone> ::core::clone::Clone for PunchrServiceClient<T> {
            #[inline]
            fn clone(&self) -> PunchrServiceClient<T> {
                match *self {
                    PunchrServiceClient {
                        inner: ref __self_0_0,
                    } => PunchrServiceClient {
                        inner: ::core::clone::Clone::clone(&(*__self_0_0)),
                    },
                }
            }
        }
        impl PunchrServiceClient<tonic::transport::Channel> {
            /// Attempt to create a new client by connecting to a given endpoint.
            pub async fn connect<D>(dst: D) -> Result<Self, tonic::transport::Error>
            where
                D: std::convert::TryInto<tonic::transport::Endpoint>,
                D::Error: Into<StdError>,
            {
                let conn = tonic::transport::Endpoint::new(dst)?.connect().await?;
                Ok(Self::new(conn))
            }
        }
        impl<T> PunchrServiceClient<T>
        where
            T: tonic::client::GrpcService<tonic::body::BoxBody>,
            T::ResponseBody: Body + Send + 'static,
            T::Error: Into<StdError>,
            <T::ResponseBody as Body>::Error: Into<StdError> + Send,
        {
            pub fn new(inner: T) -> Self {
                let inner = tonic::client::Grpc::new(inner);
                Self { inner }
            }
            pub fn with_interceptor<F>(
                inner: T,
                interceptor: F,
            ) -> PunchrServiceClient<InterceptedService<T, F>>
            where
                F: tonic::service::Interceptor,
                T: tonic::codegen::Service<
                    http::Request<tonic::body::BoxBody>,
                    Response = http::Response<
                        <T as tonic::client::GrpcService<tonic::body::BoxBody>>::ResponseBody,
                    >,
                >,
                <T as tonic::codegen::Service<http::Request<tonic::body::BoxBody>>>::Error:
                    Into<StdError> + Send + Sync,
            {
                PunchrServiceClient::new(InterceptedService::new(inner, interceptor))
            }
            /// Compress requests with `gzip`.
            ///
            /// This requires the server to support it otherwise it might respond with an
            /// error.
            pub fn send_gzip(mut self) -> Self {
                self.inner = self.inner.send_gzip();
                self
            }
            /// Enable decompressing responses with `gzip`.
            pub fn accept_gzip(mut self) -> Self {
                self.inner = self.inner.accept_gzip();
                self
            }
            /// Register takes punchr client information and saves them to the database.
            /// This should be called upon start of a client.
            pub async fn register(
                &mut self,
                request: impl tonic::IntoRequest<super::RegisterRequest>,
            ) -> Result<tonic::Response<super::RegisterResponse>, tonic::Status> {
                self.inner.ready().await.map_err(|e| {
                    tonic::Status::new(tonic::Code::Unknown, {
                        let res = ::alloc::fmt::format(::core::fmt::Arguments::new_v1(
                            &["Service was not ready: "],
                            &[::core::fmt::ArgumentV1::new_display(&e.into())],
                        ));
                        res
                    })
                })?;
                let codec = tonic::codec::ProstCodec::default();
                let path = http::uri::PathAndQuery::from_static("/PunchrService/Register");
                self.inner.unary(request.into_request(), path, codec).await
            }
            /// GetAddrInfo returns peer address information that should be used to attempt
            /// a hole punch. Clients should call this endpoint periodically.
            pub async fn get_addr_info(
                &mut self,
                request: impl tonic::IntoRequest<super::GetAddrInfoRequest>,
            ) -> Result<tonic::Response<super::GetAddrInfoResponse>, tonic::Status> {
                self.inner.ready().await.map_err(|e| {
                    tonic::Status::new(tonic::Code::Unknown, {
                        let res = ::alloc::fmt::format(::core::fmt::Arguments::new_v1(
                            &["Service was not ready: "],
                            &[::core::fmt::ArgumentV1::new_display(&e.into())],
                        ));
                        res
                    })
                })?;
                let codec = tonic::codec::ProstCodec::default();
                let path = http::uri::PathAndQuery::from_static("/PunchrService/GetAddrInfo");
                self.inner.unary(request.into_request(), path, codec).await
            }
            /// TrackHolePunch takes measurement data from the client and persists them in the database
            pub async fn track_hole_punch(
                &mut self,
                request: impl tonic::IntoRequest<super::TrackHolePunchRequest>,
            ) -> Result<tonic::Response<super::TrackHolePunchResponse>, tonic::Status> {
                self.inner.ready().await.map_err(|e| {
                    tonic::Status::new(tonic::Code::Unknown, {
                        let res = ::alloc::fmt::format(::core::fmt::Arguments::new_v1(
                            &["Service was not ready: "],
                            &[::core::fmt::ArgumentV1::new_display(&e.into())],
                        ));
                        res
                    })
                })?;
                let codec = tonic::codec::ProstCodec::default();
                let path = http::uri::PathAndQuery::from_static("/PunchrService/TrackHolePunch");
                self.inner.unary(request.into_request(), path, codec).await
            }
        }
    }
    /// Generated server implementations.
    pub mod punchr_service_server {
        #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
        use tonic::codegen::*;
        ///Generated trait containing gRPC methods that should be implemented for use with PunchrServiceServer.
        pub trait PunchrService: Send + Sync + 'static {
            /// Register takes punchr client information and saves them to the database.
            /// This should be called upon start of a client.
            #[must_use]
            #[allow(clippy::type_complexity, clippy::type_repetition_in_bounds)]
            fn register<'life0, 'async_trait>(
                &'life0 self,
                request: tonic::Request<super::RegisterRequest>,
            ) -> ::core::pin::Pin<
                Box<
                    dyn ::core::future::Future<
                            Output = Result<
                                tonic::Response<super::RegisterResponse>,
                                tonic::Status,
                            >,
                        > + ::core::marker::Send
                        + 'async_trait,
                >,
            >
            where
                'life0: 'async_trait,
                Self: 'async_trait;
            /// GetAddrInfo returns peer address information that should be used to attempt
            /// a hole punch. Clients should call this endpoint periodically.
            #[must_use]
            #[allow(clippy::type_complexity, clippy::type_repetition_in_bounds)]
            fn get_addr_info<'life0, 'async_trait>(
                &'life0 self,
                request: tonic::Request<super::GetAddrInfoRequest>,
            ) -> ::core::pin::Pin<
                Box<
                    dyn ::core::future::Future<
                            Output = Result<
                                tonic::Response<super::GetAddrInfoResponse>,
                                tonic::Status,
                            >,
                        > + ::core::marker::Send
                        + 'async_trait,
                >,
            >
            where
                'life0: 'async_trait,
                Self: 'async_trait;
            /// TrackHolePunch takes measurement data from the client and persists them in the database
            #[must_use]
            #[allow(clippy::type_complexity, clippy::type_repetition_in_bounds)]
            fn track_hole_punch<'life0, 'async_trait>(
                &'life0 self,
                request: tonic::Request<super::TrackHolePunchRequest>,
            ) -> ::core::pin::Pin<
                Box<
                    dyn ::core::future::Future<
                            Output = Result<
                                tonic::Response<super::TrackHolePunchResponse>,
                                tonic::Status,
                            >,
                        > + ::core::marker::Send
                        + 'async_trait,
                >,
            >
            where
                'life0: 'async_trait,
                Self: 'async_trait;
        }
        pub struct PunchrServiceServer<T: PunchrService> {
            inner: _Inner<T>,
            accept_compression_encodings: (),
            send_compression_encodings: (),
        }
        #[automatically_derived]
        #[allow(unused_qualifications)]
        impl<T: ::core::fmt::Debug + PunchrService> ::core::fmt::Debug for PunchrServiceServer<T> {
            fn fmt(&self, f: &mut ::core::fmt::Formatter) -> ::core::fmt::Result {
                match *self {
                    PunchrServiceServer {
                        inner: ref __self_0_0,
                        accept_compression_encodings: ref __self_0_1,
                        send_compression_encodings: ref __self_0_2,
                    } => {
                        let debug_trait_builder =
                            &mut ::core::fmt::Formatter::debug_struct(f, "PunchrServiceServer");
                        let _ = ::core::fmt::DebugStruct::field(
                            debug_trait_builder,
                            "inner",
                            &&(*__self_0_0),
                        );
                        let _ = ::core::fmt::DebugStruct::field(
                            debug_trait_builder,
                            "accept_compression_encodings",
                            &&(*__self_0_1),
                        );
                        let _ = ::core::fmt::DebugStruct::field(
                            debug_trait_builder,
                            "send_compression_encodings",
                            &&(*__self_0_2),
                        );
                        ::core::fmt::DebugStruct::finish(debug_trait_builder)
                    }
                }
            }
        }
        struct _Inner<T>(Arc<T>);
        impl<T: PunchrService> PunchrServiceServer<T> {
            pub fn new(inner: T) -> Self {
                let inner = Arc::new(inner);
                let inner = _Inner(inner);
                Self {
                    inner,
                    accept_compression_encodings: Default::default(),
                    send_compression_encodings: Default::default(),
                }
            }
            pub fn with_interceptor<F>(inner: T, interceptor: F) -> InterceptedService<Self, F>
            where
                F: tonic::service::Interceptor,
            {
                InterceptedService::new(Self::new(inner), interceptor)
            }
        }
        impl<T, B> tonic::codegen::Service<http::Request<B>> for PunchrServiceServer<T>
        where
            T: PunchrService,
            B: Body + Send + 'static,
            B::Error: Into<StdError> + Send + 'static,
        {
            type Response = http::Response<tonic::body::BoxBody>;
            type Error = Never;
            type Future = BoxFuture<Self::Response, Self::Error>;
            fn poll_ready(&mut self, _cx: &mut Context<'_>) -> Poll<Result<(), Self::Error>> {
                Poll::Ready(Ok(()))
            }
            fn call(&mut self, req: http::Request<B>) -> Self::Future {
                let inner = self.inner.clone();
                match req.uri().path() {
                    "/PunchrService/Register" => {
                        #[allow(non_camel_case_types)]
                        struct RegisterSvc<T: PunchrService>(pub Arc<T>);
                        impl<T: PunchrService> tonic::server::UnaryService<super::RegisterRequest> for RegisterSvc<T> {
                            type Response = super::RegisterResponse;
                            type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                            fn call(
                                &mut self,
                                request: tonic::Request<super::RegisterRequest>,
                            ) -> Self::Future {
                                let inner = self.0.clone();
                                let fut = async move { (*inner).register(request).await };
                                Box::pin(fut)
                            }
                        }
                        let accept_compression_encodings = self.accept_compression_encodings;
                        let send_compression_encodings = self.send_compression_encodings;
                        let inner = self.inner.clone();
                        let fut = async move {
                            let inner = inner.0;
                            let method = RegisterSvc(inner);
                            let codec = tonic::codec::ProstCodec::default();
                            let mut grpc = tonic::server::Grpc::new(codec)
                                .apply_compression_config(
                                    accept_compression_encodings,
                                    send_compression_encodings,
                                );
                            let res = grpc.unary(method, req).await;
                            Ok(res)
                        };
                        Box::pin(fut)
                    }
                    "/PunchrService/GetAddrInfo" => {
                        #[allow(non_camel_case_types)]
                        struct GetAddrInfoSvc<T: PunchrService>(pub Arc<T>);
                        impl<T: PunchrService>
                            tonic::server::UnaryService<super::GetAddrInfoRequest>
                            for GetAddrInfoSvc<T>
                        {
                            type Response = super::GetAddrInfoResponse;
                            type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                            fn call(
                                &mut self,
                                request: tonic::Request<super::GetAddrInfoRequest>,
                            ) -> Self::Future {
                                let inner = self.0.clone();
                                let fut = async move { (*inner).get_addr_info(request).await };
                                Box::pin(fut)
                            }
                        }
                        let accept_compression_encodings = self.accept_compression_encodings;
                        let send_compression_encodings = self.send_compression_encodings;
                        let inner = self.inner.clone();
                        let fut = async move {
                            let inner = inner.0;
                            let method = GetAddrInfoSvc(inner);
                            let codec = tonic::codec::ProstCodec::default();
                            let mut grpc = tonic::server::Grpc::new(codec)
                                .apply_compression_config(
                                    accept_compression_encodings,
                                    send_compression_encodings,
                                );
                            let res = grpc.unary(method, req).await;
                            Ok(res)
                        };
                        Box::pin(fut)
                    }
                    "/PunchrService/TrackHolePunch" => {
                        #[allow(non_camel_case_types)]
                        struct TrackHolePunchSvc<T: PunchrService>(pub Arc<T>);
                        impl<T: PunchrService>
                            tonic::server::UnaryService<super::TrackHolePunchRequest>
                            for TrackHolePunchSvc<T>
                        {
                            type Response = super::TrackHolePunchResponse;
                            type Future = BoxFuture<tonic::Response<Self::Response>, tonic::Status>;
                            fn call(
                                &mut self,
                                request: tonic::Request<super::TrackHolePunchRequest>,
                            ) -> Self::Future {
                                let inner = self.0.clone();
                                let fut = async move { (*inner).track_hole_punch(request).await };
                                Box::pin(fut)
                            }
                        }
                        let accept_compression_encodings = self.accept_compression_encodings;
                        let send_compression_encodings = self.send_compression_encodings;
                        let inner = self.inner.clone();
                        let fut = async move {
                            let inner = inner.0;
                            let method = TrackHolePunchSvc(inner);
                            let codec = tonic::codec::ProstCodec::default();
                            let mut grpc = tonic::server::Grpc::new(codec)
                                .apply_compression_config(
                                    accept_compression_encodings,
                                    send_compression_encodings,
                                );
                            let res = grpc.unary(method, req).await;
                            Ok(res)
                        };
                        Box::pin(fut)
                    }
                    _ => Box::pin(async move {
                        Ok(http::Response::builder()
                            .status(200)
                            .header("grpc-status", "12")
                            .header("content-type", "application/grpc")
                            .body(empty_body())
                            .unwrap())
                    }),
                }
            }
        }
        impl<T: PunchrService> Clone for PunchrServiceServer<T> {
            fn clone(&self) -> Self {
                let inner = self.inner.clone();
                Self {
                    inner,
                    accept_compression_encodings: self.accept_compression_encodings,
                    send_compression_encodings: self.send_compression_encodings,
                }
            }
        }
        impl<T: PunchrService> Clone for _Inner<T> {
            fn clone(&self) -> Self {
                Self(self.0.clone())
            }
        }
        impl<T: std::fmt::Debug> std::fmt::Debug for _Inner<T> {
            fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                f.write_fmt(::core::fmt::Arguments::new_v1(
                    &[""],
                    &[::core::fmt::ArgumentV1::new_debug(&self.0)],
                ))
            }
        }
        impl<T: PunchrService> tonic::transport::NamedService for PunchrServiceServer<T> {
            const NAME: &'static str = "PunchrService";
        }
    }
}
