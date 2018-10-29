package endpoints

import (
	"encoding/json"
	"fmt"
	"sync"
)

const endpointsJson = "{" +
	"  \"products\":[" +
	"  {" +
	"    \"code\": \"aegis\"," +
	"    \"document_id\": \"28449\"," +
	"    \"location_service_code\": \"vipaegis\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"aegis.cn-hangzhou.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"alidns\"," +
	"    \"document_id\": \"29739\"," +
	"    \"location_service_code\": \"alidns\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"alidns.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"arms\"," +
	"    \"document_id\": \"42924\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": [ {" +
	"       \"region\": \"ap-southeast-1\"," +
	"       \"endpoint\": \"arms.ap-southeast-1.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-beijing\"," +
	"       \"endpoint\": \"arms.cn-beijing.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-hangzhou\"," +
	"       \"endpoint\": \"arms.cn-hangzhou.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-hongkong\"," +
	"       \"endpoint\": \"arms.cn-hongkong.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-qingdao\"," +
	"       \"endpoint\": \"arms.cn-qingdao.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-shanghai\"," +
	"       \"endpoint\": \"arms.cn-shanghai.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-shenzhen\"," +
	"       \"endpoint\": \"arms.cn-shenzhen.aliyuncs.com\"" +
	"    }]," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"arms.[RegionId].aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"batchcompute\"," +
	"    \"document_id\": \"44717\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": [ {" +
	"       \"region\": \"ap-southeast-1\"," +
	"       \"endpoint\": \"batchcompute.ap-southeast-1.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-beijing\"," +
	"       \"endpoint\": \"batchcompute.cn-beijing.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-hangzhou\"," +
	"       \"endpoint\": \"batchcompute.cn-hangzhou.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-huhehaote\"," +
	"       \"endpoint\": \"batchcompute.cn-huhehaote.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-qingdao\"," +
	"       \"endpoint\": \"batchcompute.cn-qingdao.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-shanghai\"," +
	"       \"endpoint\": \"batchcompute.cn-shanghai.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-shenzhen\"," +
	"       \"endpoint\": \"batchcompute.cn-shenzhen.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-zhangjiakou\"," +
	"       \"endpoint\": \"batchcompute.cn-zhangjiakou.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"us-west-1\"," +
	"       \"endpoint\": \"batchcompute.us-west-1.aliyuncs.com\"" +
	"    }]," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"batchcompute.[RegionId].aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"ccc\"," +
	"    \"document_id\": \"63027\"," +
	"    \"location_service_code\": \"ccc\"," +
	"    \"regional_endpoints\": [ {" +
	"       \"region\": \"cn-hangzhou\"," +
	"       \"endpoint\": \"ccc.cn-hangzhou.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-shanghai\"," +
	"       \"endpoint\": \"ccc.cn-shanghai.aliyuncs.com\"" +
	"    }]," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"ccc.[RegionId].aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"cdn\"," +
	"    \"document_id\": \"27148\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"cdn.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"cds\"," +
	"    \"document_id\": \"62887\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"cds.cn-beijing.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"chatbot\"," +
	"    \"document_id\": \"60760\"," +
	"    \"location_service_code\": \"beebot\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"chatbot.[RegionId].aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"cloudapi\"," +
	"    \"document_id\": \"43590\"," +
	"    \"location_service_code\": \"apigateway\"," +
	"    \"regional_endpoints\": [ {" +
	"       \"region\": \"ap-northeast-1\"," +
	"       \"endpoint\": \"apigateway.ap-northeast-1.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"us-west-1\"," +
	"       \"endpoint\": \"apigateway.us-west-1.aliyuncs.com\"" +
	"    }]," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"apigateway.[RegionId].aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"cloudauth\"," +
	"    \"document_id\": \"60687\"," +
	"    \"location_service_code\": \"cloudauth\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"cloudauth.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"cloudphoto\"," +
	"    \"document_id\": \"59902\"," +
	"    \"location_service_code\": \"cloudphoto\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"cloudphoto.[RegionId].aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"cloudwf\"," +
	"    \"document_id\": \"58111\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"cloudwf.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"cms\"," +
	"    \"document_id\": \"28615\"," +
	"    \"location_service_code\": \"cms\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"cr\"," +
	"    \"document_id\": \"60716\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"cr.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"cs\"," +
	"    \"document_id\": \"26043\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"cs.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"csb\"," +
	"    \"document_id\": \"64837\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": [ {" +
	"       \"region\": \"cn-beijing\"," +
	"       \"endpoint\": \"csb.cn-beijing.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-hangzhou\"," +
	"       \"endpoint\": \"csb.cn-hangzhou.aliyuncs.com\"" +
	"    }]," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"csb.[RegionId].aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"dds\"," +
	"    \"document_id\": \"61715\"," +
	"    \"location_service_code\": \"dds\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"mongodb.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"mongodb.[RegionId].aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"dm\"," +
	"    \"document_id\": \"29434\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": [ {" +
	"       \"region\": \"ap-southeast-1\"," +
	"       \"endpoint\": \"dm.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"ap-southeast-2\"," +
	"       \"endpoint\": \"dm.ap-southeast-2.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-beijing\"," +
	"       \"endpoint\": \"dm.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-hangzhou\"," +
	"       \"endpoint\": \"dm.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-hongkong\"," +
	"       \"endpoint\": \"dm.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-qingdao\"," +
	"       \"endpoint\": \"dm.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-shanghai\"," +
	"       \"endpoint\": \"dm.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"cn-shenzhen\"," +
	"       \"endpoint\": \"dm.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"us-east-1\"," +
	"       \"endpoint\": \"dm.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"us-west-1\"," +
	"       \"endpoint\": \"dm.aliyuncs.com\"" +
	"    }]," +
	"    \"global_endpoint\": \"dm.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"dm.[RegionId].aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"domain\"," +
	"    \"document_id\": \"42875\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"domain.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"domain.aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"domain-intl\"," +
	"    \"document_id\": \"\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"domain-intl.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"domain-intl.aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"drds\"," +
	"    \"document_id\": \"51111\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"drds.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"drds.aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"ecs\"," +
	"    \"document_id\": \"25484\"," +
	"    \"location_service_code\": \"ecs\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"emr\"," +
	"    \"document_id\": \"28140\"," +
	"    \"location_service_code\": \"emr\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"emr.[RegionId].aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"ess\"," +
	"    \"document_id\": \"25925\"," +
	"    \"location_service_code\": \"ess\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"ess.[RegionId].aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"green\"," +
	"    \"document_id\": \"28427\"," +
	"    \"location_service_code\": \"green\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"green.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"hpc\"," +
	"    \"document_id\": \"35201\"," +
	"    \"location_service_code\": \"hpc\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"hpc.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"httpdns\"," +
	"    \"document_id\": \"52679\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"httpdns-api.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"iot\"," +
	"    \"document_id\": \"30557\"," +
	"    \"location_service_code\": \"iot\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"iot.[RegionId].aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"itaas\"," +
	"    \"document_id\": \"55759\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"itaas.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"jaq\"," +
	"    \"document_id\": \"35037\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"jaq.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"live\"," +
	"    \"document_id\": \"48207\"," +
	"    \"location_service_code\": \"live\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"live.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"mts\"," +
	"    \"document_id\": \"29212\"," +
	"    \"location_service_code\": \"mts\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"nas\"," +
	"    \"document_id\": \"62598\"," +
	"    \"location_service_code\": \"nas\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"ons\"," +
	"    \"document_id\": \"44416\"," +
	"    \"location_service_code\": \"ons\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"polardb\"," +
	"    \"document_id\": \"58764\"," +
	"    \"location_service_code\": \"polardb\"," +
	"    \"regional_endpoints\": [ {" +
	"       \"region\": \"ap-south-1\"," +
	"       \"endpoint\": \"polardb.ap-south-1.aliyuncs.com\"" +
	"    }, {" +
	"       \"region\": \"ap-southeast-5\"," +
	"       \"endpoint\": \"polardb.ap-southeast-5.aliyuncs.com\"" +
	"    }]," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"polardb.aliyuncs.com\"" +
	"  }," +
	"  {" +
	"    \"code\": \"push\"," +
	"    \"document_id\": \"30074\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"cloudpush.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"qualitycheck\"," +
	"    \"document_id\": \"50807\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": [ {" +
	"       \"region\": \"cn-hangzhou\"," +
	"       \"endpoint\": \"qualitycheck.cn-hangzhou.aliyuncs.com\"" +
	"    }]," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"r-kvstore\"," +
	"    \"document_id\": \"60831\"," +
	"    \"location_service_code\": \"redisa\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"ram\"," +
	"    \"document_id\": \"28672\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"ram.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"rds\"," +
	"    \"document_id\": \"26223\"," +
	"    \"location_service_code\": \"rds\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"ros\"," +
	"    \"document_id\": \"28899\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"ros.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"sas-api\"," +
	"    \"document_id\": \"28498\"," +
	"    \"location_service_code\": \"sas\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"slb\"," +
	"    \"document_id\": \"27565\"," +
	"    \"location_service_code\": \"slb\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"sts\"," +
	"    \"document_id\": \"28756\"," +
	"    \"location_service_code\": \"\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"sts.aliyuncs.com\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"vod\"," +
	"    \"document_id\": \"60574\"," +
	"    \"location_service_code\": \"vod\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"vpc\"," +
	"    \"document_id\": \"34962\"," +
	"    \"location_service_code\": \"vpc\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }," +
	"  {" +
	"    \"code\": \"waf\"," +
	"    \"document_id\": \"62847\"," +
	"    \"location_service_code\": \"waf\"," +
	"    \"regional_endpoints\": []," +
	"    \"global_endpoint\": \"\"," +
	"    \"regional_endpoint_pattern\": \"\"" +
	"  }]" +
	"}"

var initOnce sync.Once
var data interface{}

func getEndpointConfigData() interface{} {
	initOnce.Do(func() {
		err := json.Unmarshal([]byte(endpointsJson), &data)
		if err != nil {
			fmt.Println("init endpoint config data failed.", err)
		}
	})
	return data
}
