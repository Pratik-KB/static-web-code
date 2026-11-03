package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type Endpoint struct {
	URL    string
	Method string
}

type Target struct {
	URL             string
	Method          string
	ExpandedFromAny bool
}

type Result struct {
	URL                   string            `json:"url"`
	Method                string            `json:"method"`
	TimestampUTC          string            `json:"timestamp_utc"`
	ExpandedFromAny       bool              `json:"expanded_from_any"`
	Skipped               bool              `json:"skipped"`
	Reason                string            `json:"reason,omitempty"`
	StatusCode            int               `json:"status_code,omitempty"`
	ElapsedMs             float64           `json:"elapsed_ms,omitempty"`
	ResponseHeaders       map[string]string `json:"response_headers,omitempty"`
	ResponseBody          string            `json:"response_body,omitempty"`
	ResponseBodyTruncated bool              `json:"response_body_truncated,omitempty"`
	ResponseSizeBytes     int               `json:"response_size_bytes,omitempty"`
	HTTPError             bool              `json:"http_error,omitempty"`
	Error                 string            `json:"error,omitempty"`
	ErrorMessage          string            `json:"error_message,omitempty"`
}

const endpointTSV = `
Url	Method
https://dblaccountsapi.kbnbfc.in/v1/login	POST
https://dblaccountsapi.kbnbfc.in/v1/login	PUT
https://dblaccountsapi.kbnbfc.in/v1/logout	POST
https://dblaccountsapi.kbnbfc.in/v1/logout/app	POST
https://dblaccountsapi.kbnbfc.in/v1/passcode	PUT
https://dblapi.kbnbfc.in/admin/aadhaarmask	POST
https://dblapi.kbnbfc.in/admin/aadhaarwarning	GET
https://dblapi.kbnbfc.in/admin/aadhaarwarning	POST
https://dblapi.kbnbfc.in/admin/aadhaarwarning	PUT
https://dblapi.kbnbfc.in/admin/accountdeletioncsv/history	GET
https://dblapi.kbnbfc.in/admin/accountdeletioncsv/import	POST
https://dblapi.kbnbfc.in/admin/accountdeletioncsv/presignedurl	GET
https://dblapi.kbnbfc.in/admin/admin	ANY
https://dblapi.kbnbfc.in/admin/admin/agent	GET
https://dblapi.kbnbfc.in/admin/admin/agent	POST
https://dblapi.kbnbfc.in/admin/admin/agent/bucket	GET
https://dblapi.kbnbfc.in/admin/admin/agent/markabsent	PUT
https://dblapi.kbnbfc.in/admin/admin/customer/rekycenable	POST
https://dblapi.kbnbfc.in/admin/admin/customer/soahistory	GET
https://dblapi.kbnbfc.in/admin/api/13175	GET
https://dblapi.kbnbfc.in/admin/api/38122	POST
https://dblapi.kbnbfc.in/admin/apps	GET
https://dblapi.kbnbfc.in/admin/autodebitGrid	GET
https://dblapi.kbnbfc.in/admin/balancemonitor/edit	POST
https://dblapi.kbnbfc.in/admin/balancemonitor/export	GET
https://dblapi.kbnbfc.in/admin/balancemonitor/grid	GET
https://dblapi.kbnbfc.in/admin/banks	ANY
https://dblapi.kbnbfc.in/admin/batch	ANY
https://dblapi.kbnbfc.in/admin/batch/history	GET
https://dblapi.kbnbfc.in/admin/bp/bandproperties	GET
https://dblapi.kbnbfc.in/admin/bp/grid	GET
https://dblapi.kbnbfc.in/admin/bp/newreq	POST
https://dblapi.kbnbfc.in/admin/bp/products	GET
https://dblapi.kbnbfc.in/admin/bp/updatereq	PUT
https://dblapi.kbnbfc.in/admin/bp/view	GET
https://dblapi.kbnbfc.in/admin/callNow	ANY
https://dblapi.kbnbfc.in/admin/cashback	GET
https://dblapi.kbnbfc.in/admin/collection/add	PUT
https://dblapi.kbnbfc.in/admin/collection/agent/summary	GET
https://dblapi.kbnbfc.in/admin/collection/allocation	GET
https://dblapi.kbnbfc.in/admin/collection/allocation/agents	GET
https://dblapi.kbnbfc.in/admin/collection/allocation/agents	POST
https://dblapi.kbnbfc.in/admin/collection/allocation/agents	PUT
https://dblapi.kbnbfc.in/admin/collection/allocation/agents/grid	GET
https://dblapi.kbnbfc.in/admin/collection/allocation/agents/import	GET
https://dblapi.kbnbfc.in/admin/collection/allocation/agents/import	POST
https://dblapi.kbnbfc.in/admin/collection/allocation/agents/markabsent	POST
https://dblapi.kbnbfc.in/admin/collection/allocation/config	GET
https://dblapi.kbnbfc.in/admin/collection/allocation/disposition	POST
https://dblapi.kbnbfc.in/admin/collection/allocation/import	GET
https://dblapi.kbnbfc.in/admin/collection/allocation/import	POST
https://dblapi.kbnbfc.in/admin/collection/allocation/presignedurl	GET
https://dblapi.kbnbfc.in/admin/collection/allocation/reallocate	GET
https://dblapi.kbnbfc.in/admin/collection/allocation/reallocate/callback	POST
https://dblapi.kbnbfc.in/admin/collection/automation/ptp	POST
https://dblapi.kbnbfc.in/admin/collection/automation/ptp/history	GET
https://dblapi.kbnbfc.in/admin/collection/call	ANY
https://dblapi.kbnbfc.in/admin/collection/cancelnach/grid	GET
https://dblapi.kbnbfc.in/admin/collection/cancelnach/upload	PUT
https://dblapi.kbnbfc.in/admin/collection/disableAD/grid	GET
https://dblapi.kbnbfc.in/admin/collection/disableAD/upload	PUT
https://dblapi.kbnbfc.in/admin/collection/edit	POST
https://dblapi.kbnbfc.in/admin/collection/grid	GET
https://dblapi.kbnbfc.in/admin/collection/linknach/grid	GET
https://dblapi.kbnbfc.in/admin/collection/linknach/upload	PUT
https://dblapi.kbnbfc.in/admin/collection/loanbucket	GET
https://dblapi.kbnbfc.in/admin/collection/loanbucket	POST
https://dblapi.kbnbfc.in/admin/collection/loanbucket	PUT
https://dblapi.kbnbfc.in/admin/collection/mandatestatus/grid	GET
https://dblapi.kbnbfc.in/admin/collection/mandatestatus/upload	PUT
https://dblapi.kbnbfc.in/admin/collection/mobile/grid	GET
https://dblapi.kbnbfc.in/admin/collection/nachautomandate/add	PUT
https://dblapi.kbnbfc.in/admin/collection/nachautomandate/grid	GET
https://dblapi.kbnbfc.in/admin/collection/notification/callback	POST
https://dblapi.kbnbfc.in/admin/collection/notification/import	POST
https://dblapi.kbnbfc.in/admin/collection/notification/presignedurl	GET
https://dblapi.kbnbfc.in/admin/collection/partner/disposition/18759	POST
https://dblapi.kbnbfc.in/admin/collection/summary	GET
https://dblapi.kbnbfc.in/admin/collection/tracking/virtualnumber/edit	POST
https://dblapi.kbnbfc.in/admin/collection/tracking/virtualnumber/grid	GET
https://dblapi.kbnbfc.in/admin/collection/tracking/virtualnumber/import	POST
https://dblapi.kbnbfc.in/admin/collection/view	GET
https://dblapi.kbnbfc.in/admin/collection/36408	GET
https://dblapi.kbnbfc.in/admin/collectionpartner/grid/creditmate	GET
https://dblapi.kbnbfc.in/admin/collectionpartner/uploadcase/creditmate	POST
https://dblapi.kbnbfc.in/admin/comment	ANY
https://dblapi.kbnbfc.in/admin/company	GET
https://dblapi.kbnbfc.in/admin/coupon/code	ANY
https://dblapi.kbnbfc.in/admin/coupon/code/assigned	POST
https://dblapi.kbnbfc.in/admin/coupon/code/bandlist	GET
https://dblapi.kbnbfc.in/admin/coupon/code/existing	GET
https://dblapi.kbnbfc.in/admin/coupon/code/grid	GET
https://dblapi.kbnbfc.in/admin/coupon/code/history_grid	GET
https://dblapi.kbnbfc.in/admin/coupon/code/remaining	GET
https://dblapi.kbnbfc.in/admin/coupon/code/upload	POST
https://dblapi.kbnbfc.in/admin/coupon/history/grid	GET
https://dblapi.kbnbfc.in/admin/csv	ANY
https://dblapi.kbnbfc.in/admin/customer/accountdelete	POST
https://dblapi.kbnbfc.in/admin/customer/adriskchecklist	GET
https://dblapi.kbnbfc.in/admin/customer/adriskchecklistsubmit	POST
https://dblapi.kbnbfc.in/admin/customer/adriskhistory	GET
https://dblapi.kbnbfc.in/admin/customer/digiKycDocs	GET
https://dblapi.kbnbfc.in/admin/customer/digiKycRiskChecklist	GET
https://dblapi.kbnbfc.in/admin/customer/digiKycRiskHistory	GET
https://dblapi.kbnbfc.in/admin/customer/digiKycRiskchecklistSubmit	POST
https://dblapi.kbnbfc.in/admin/customer/insurance	GET
https://dblapi.kbnbfc.in/admin/customer/loanproductsinfo	GET
https://dblapi.kbnbfc.in/admin/customer/pdriskchecklist	GET
https://dblapi.kbnbfc.in/admin/customer/pdriskchecklistsubmit	POST
https://dblapi.kbnbfc.in/admin/customer/pdriskhistory	GET
https://dblapi.kbnbfc.in/admin/customer/rekycenable	POST
https://dblapi.kbnbfc.in/admin/customer/35354	ANY
https://dblapi.kbnbfc.in/admin/customersupport/callback/grid	GET
https://dblapi.kbnbfc.in/admin/customersupport/callback/grid	PUT
https://dblapi.kbnbfc.in/admin/customersupport/email/grid	GET
https://dblapi.kbnbfc.in/admin/customersupport/email/import	POST
https://dblapi.kbnbfc.in/admin/dbl-callback/exotel	POST
https://dblapi.kbnbfc.in/admin/digitalkyc/agent/bulkuploadhistory	GET
https://dblapi.kbnbfc.in/admin/digitalkyc/agent/import	GET
https://dblapi.kbnbfc.in/admin/digitalkyc/agent/import	POST
https://dblapi.kbnbfc.in/admin/disableAutodebit	POST
https://dblapi.kbnbfc.in/admin/ecom/bandprop/grid	GET
https://dblapi.kbnbfc.in/admin/ecom/bandprop/history	GET
https://dblapi.kbnbfc.in/admin/ecom/banner	ANY
https://dblapi.kbnbfc.in/admin/ecom/banner/grid	GET
https://dblapi.kbnbfc.in/admin/ecom/banner/types	GET
https://dblapi.kbnbfc.in/admin/ecom/category/grid	GET
https://dblapi.kbnbfc.in/admin/ecom/category/list	GET
https://dblapi.kbnbfc.in/admin/ecom/giftcard/gridone	GET
https://dblapi.kbnbfc.in/admin/ecom/giftcard/gridtwo	GET
https://dblapi.kbnbfc.in/admin/ecom/partner/grid	GET
https://dblapi.kbnbfc.in/admin/ecom/product	ANY
https://dblapi.kbnbfc.in/admin/ecom/product/csv/grid	GET
https://dblapi.kbnbfc.in/admin/ecom/product/grid	GET
https://dblapi.kbnbfc.in/admin/ecom/product/import	POST
https://dblapi.kbnbfc.in/admin/ecom/product/partnerlist	GET
https://dblapi.kbnbfc.in/admin/editablepaymentlink	GET
https://dblapi.kbnbfc.in/admin/editablepaymentlink	POST
https://dblapi.kbnbfc.in/admin/editablepaymentlink/resend	POST
https://dblapi.kbnbfc.in/admin/enach	ANY
https://dblapi.kbnbfc.in/admin/enachifsc/config	GET
https://dblapi.kbnbfc.in/admin/enachifsc/config/history	GET
https://dblapi.kbnbfc.in/admin/excesspaymentrefund/upload	POST
https://dblapi.kbnbfc.in/admin/exitEmployee	GET
https://dblapi.kbnbfc.in/admin/exitEmployee	PUT
https://dblapi.kbnbfc.in/admin/fcl/history	GET
https://dblapi.kbnbfc.in/admin/fcl/import	POST
https://dblapi.kbnbfc.in/admin/fcl/presigned	GET
https://dblapi.kbnbfc.in/admin/hardlegal	ANY
https://dblapi.kbnbfc.in/admin/hardlegal/delivery	GET
https://dblapi.kbnbfc.in/admin/hardlegal/delivery	POST
https://dblapi.kbnbfc.in/admin/hardlegal/delivery/presignedurl	GET
https://dblapi.kbnbfc.in/admin/hardlegal/docketid/count	GET
https://dblapi.kbnbfc.in/admin/hardlegal/docketid/grid	GET
https://dblapi.kbnbfc.in/admin/hardlegal/docketid/history	GET
https://dblapi.kbnbfc.in/admin/hardlegal/docketid/import	POST
https://dblapi.kbnbfc.in/admin/hardlegal/docketid/update	POST
https://dblapi.kbnbfc.in/admin/home/notify/paymentTransaction	GET
https://dblapi.kbnbfc.in/admin/home/notify/waiver	GET
https://dblapi.kbnbfc.in/admin/iifl/bucket/log/grid	GET
https://dblapi.kbnbfc.in/admin/iifl/bucket/log/import	POST
https://dblapi.kbnbfc.in/admin/incomeverify/bank	GET
https://dblapi.kbnbfc.in/admin/incomeverify/bank	POST
https://dblapi.kbnbfc.in/admin/incomeverify/eligibility	GET
https://dblapi.kbnbfc.in/admin/incomeverify/incomeverifybanks	GET
https://dblapi.kbnbfc.in/admin/incomeverify/ivbank	GET
https://dblapi.kbnbfc.in/admin/incomeverify/presignedurl	GET
https://dblapi.kbnbfc.in/admin/incomeverify/statement	GET
https://dblapi.kbnbfc.in/admin/incomeverify/statement	POST
https://dblapi.kbnbfc.in/admin/incomeverify/updatebankdetails	POST
https://dblapi.kbnbfc.in/admin/incomeverify/updatebanks	PUT
https://dblapi.kbnbfc.in/admin/incomeverify/upload	POST
https://dblapi.kbnbfc.in/admin/insurer/all	GET
https://dblapi.kbnbfc.in/admin/insurer/create	POST
https://dblapi.kbnbfc.in/admin/insurer/details	GET
https://dblapi.kbnbfc.in/admin/insurer/getInsurerTypes	GET
https://dblapi.kbnbfc.in/admin/insurer/list	GET
https://dblapi.kbnbfc.in/admin/insurer/update	PUT
https://dblapi.kbnbfc.in/admin/interest	PUT
https://dblapi.kbnbfc.in/admin/iv/config	GET
https://dblapi.kbnbfc.in/admin/key	ANY
https://dblapi.kbnbfc.in/admin/lead-vendors	GET
https://dblapi.kbnbfc.in/admin/legalnotice/visit	GET
https://dblapi.kbnbfc.in/admin/legaltemplates/create	POST
https://dblapi.kbnbfc.in/admin/legaltemplates/details	GET
https://dblapi.kbnbfc.in/admin/legaltemplates/generate	POST
https://dblapi.kbnbfc.in/admin/legaltemplates/grid	GET
https://dblapi.kbnbfc.in/admin/legaltemplates/upload	POST
https://dblapi.kbnbfc.in/admin/linkVendorNach	POST
https://dblapi.kbnbfc.in/admin/logout	POST
https://dblapi.kbnbfc.in/admin/mailhistory	GET
https://dblapi.kbnbfc.in/admin/makerchecker	ANY
https://dblapi.kbnbfc.in/admin/mandateCancel	POST
https://dblapi.kbnbfc.in/admin/maskList	GET
https://dblapi.kbnbfc.in/admin/maskList	POST
https://dblapi.kbnbfc.in/admin/media	GET
https://dblapi.kbnbfc.in/admin/media	POST
https://dblapi.kbnbfc.in/admin/misc/6e42e9b6c48d65b07763484ba19d19cb	GET
https://dblapi.kbnbfc.in/admin/nbfc	ANY
https://dblapi.kbnbfc.in/admin/noc	ANY
https://dblapi.kbnbfc.in/admin/ocrmanualcheck/grid	GET
https://dblapi.kbnbfc.in/admin/ocrmanualcheck/info	GET
https://dblapi.kbnbfc.in/admin/ocrmanualcheck/submit	POST
https://dblapi.kbnbfc.in/admin/offers	GET
https://dblapi.kbnbfc.in/admin/offers	PUT
https://dblapi.kbnbfc.in/admin/operations/customer/search	GET
https://dblapi.kbnbfc.in/admin/operations/interest/bulk	GET
https://dblapi.kbnbfc.in/admin/operations/interest/bulk	POST
https://dblapi.kbnbfc.in/admin/operations/interest/bulk/history	GET
https://dblapi.kbnbfc.in/admin/order	ANY
https://dblapi.kbnbfc.in/admin/order/foreclosure	GET
https://dblapi.kbnbfc.in/admin/order/installments	POST
https://dblapi.kbnbfc.in/admin/order/loanDoc	GET
https://dblapi.kbnbfc.in/admin/order/settlementEmail	GET
https://dblapi.kbnbfc.in/admin/order/waiver	GET
https://dblapi.kbnbfc.in/admin/order/waiver	POST
https://dblapi.kbnbfc.in/admin/order/waiver	PUT
https://dblapi.kbnbfc.in/admin/order/waiver/grid	GET
https://dblapi.kbnbfc.in/admin/order/waiver/import	POST
https://dblapi.kbnbfc.in/admin/order/waiver/presignedurl	GET
https://dblapi.kbnbfc.in/admin/order/13365	GET
https://dblapi.kbnbfc.in/admin/order/19339	POST
https://dblapi.kbnbfc.in/admin/othercompany/card	GET
https://dblapi.kbnbfc.in/admin/othercompany/card	POST
https://dblapi.kbnbfc.in/admin/othercompany/companies	GET
https://dblapi.kbnbfc.in/admin/othercompany/verifycompany	GET
https://dblapi.kbnbfc.in/admin/othercompany/verifycompany	POST
https://dblapi.kbnbfc.in/admin/othercompany/view	GET
https://dblapi.kbnbfc.in/admin/payment	ANY
https://dblapi.kbnbfc.in/admin/paymentlink/history	GET
https://dblapi.kbnbfc.in/admin/paymentlink/import	POST
https://dblapi.kbnbfc.in/admin/paymentlink/presignedurl	GET
https://dblapi.kbnbfc.in/admin/penalty	ANY
https://dblapi.kbnbfc.in/admin/penalty/grid	GET
https://dblapi.kbnbfc.in/admin/penalty/import	POST
https://dblapi.kbnbfc.in/admin/performance	GET
https://dblapi.kbnbfc.in/admin/permission	ANY
https://dblapi.kbnbfc.in/admin/pgwarning	GET
https://dblapi.kbnbfc.in/admin/pgwarning	POST
https://dblapi.kbnbfc.in/admin/pgwarning	PUT
https://dblapi.kbnbfc.in/admin/physical-letter/download/input	GET
https://dblapi.kbnbfc.in/admin/physical-letter/download/output	GET
https://dblapi.kbnbfc.in/admin/physical-letter/download/zip	GET
https://dblapi.kbnbfc.in/admin/physical-letter/history	GET
https://dblapi.kbnbfc.in/admin/physical-letter/upload	GET
https://dblapi.kbnbfc.in/admin/physical-letter/upload	POST
https://dblapi.kbnbfc.in/admin/pkyc/scheduling/config	GET
https://dblapi.kbnbfc.in/admin/pkyc/scheduling/config	POST
https://dblapi.kbnbfc.in/admin/pkyc/scheduling/grid	GET
https://dblapi.kbnbfc.in/admin/presignedurl	GET
https://dblapi.kbnbfc.in/admin/product	ANY
https://dblapi.kbnbfc.in/admin/product/list	GET
https://dblapi.kbnbfc.in/admin/re/config	GET
https://dblapi.kbnbfc.in/admin/reason	ANY
https://dblapi.kbnbfc.in/admin/referral	ANY
https://dblapi.kbnbfc.in/admin/repaymentnoapp	ANY
https://dblapi.kbnbfc.in/admin/search	ANY
https://dblapi.kbnbfc.in/admin/ses	ANY
https://dblapi.kbnbfc.in/admin/sms	ANY
https://dblapi.kbnbfc.in/admin/sms/approve	POST
https://dblapi.kbnbfc.in/admin/sms/custom	GET
https://dblapi.kbnbfc.in/admin/sms/custom	POST
https://dblapi.kbnbfc.in/admin/sms/schedule	POST
https://dblapi.kbnbfc.in/admin/sns	POST
https://dblapi.kbnbfc.in/admin/soa/history	GET
https://dblapi.kbnbfc.in/admin/soa/import	POST
https://dblapi.kbnbfc.in/admin/soa/presigned	GET
https://dblapi.kbnbfc.in/admin/soaRequest	POST
https://dblapi.kbnbfc.in/admin/softlegalnotice/visit	GET
https://dblapi.kbnbfc.in/admin/stamppapers	GET
https://dblapi.kbnbfc.in/admin/telesales	GET
https://dblapi.kbnbfc.in/admin/telesales	POST
https://dblapi.kbnbfc.in/admin/telesales	PUT
https://dblapi.kbnbfc.in/admin/telesales/agent	GET
https://dblapi.kbnbfc.in/admin/telesales/agent	POST
https://dblapi.kbnbfc.in/admin/telesales/agent	PUT
https://dblapi.kbnbfc.in/admin/telesales/agent/grid	GET
https://dblapi.kbnbfc.in/admin/telesales/agent/markabsent	POST
https://dblapi.kbnbfc.in/admin/telesales/allocation	GET
https://dblapi.kbnbfc.in/admin/telesales/allocation/config	GET
https://dblapi.kbnbfc.in/admin/telesales/allocation/disposition	PUT
https://dblapi.kbnbfc.in/admin/telesales/allocation/import	GET
https://dblapi.kbnbfc.in/admin/telesales/allocation/import	POST
https://dblapi.kbnbfc.in/admin/telesales/allocation/presignedurl	GET
https://dblapi.kbnbfc.in/admin/telesales/autoallocation	GET
https://dblapi.kbnbfc.in/admin/telesales/autoallocation	POST
https://dblapi.kbnbfc.in/admin/telesales/autoallocation	PUT
https://dblapi.kbnbfc.in/admin/telesales/call	POST
https://dblapi.kbnbfc.in/admin/telesales/external/allocation	GET
https://dblapi.kbnbfc.in/admin/telesales/external/allocation	PUT
https://dblapi.kbnbfc.in/admin/telesales/external/allocation/config	GET
https://dblapi.kbnbfc.in/admin/telesales/external/allocation/import	GET
https://dblapi.kbnbfc.in/admin/telesales/external/allocation/import	POST
https://dblapi.kbnbfc.in/admin/telesales/external/allocation/presignedurl	GET
https://dblapi.kbnbfc.in/admin/transaction	ANY
https://dblapi.kbnbfc.in/admin/transaction/fetchPartPrepaymentDetails	POST
https://dblapi.kbnbfc.in/admin/transaction/fetchTransactionDetails	GET
https://dblapi.kbnbfc.in/admin/transaction/foreclosureEligibility	GET
https://dblapi.kbnbfc.in/admin/txn/reversal/requests	POST
https://dblapi.kbnbfc.in/admin/txn/reversal/requests/13817/42145	PUT
https://dblapi.kbnbfc.in/admin/updatemandatesrollout/nach	POST
https://dblapi.kbnbfc.in/admin/updatemandatesrollout/routing	GET
https://dblapi.kbnbfc.in/admin/updatemandatesrollout/upi	POST
https://dblapi.kbnbfc.in/admin/user	ANY
https://dblapi.kbnbfc.in/admin/user/offline/enach	POST
https://dblapi.kbnbfc.in/admin/user/offline/verification	PUT
https://dblapi.kbnbfc.in/admin/user/osvcav	GET
https://dblapi.kbnbfc.in/admin/user/osvcav	PUT
https://dblapi.kbnbfc.in/admin/user/osvcav/grid	GET
https://dblapi.kbnbfc.in/admin/user/perfios/reset	POST
https://dblapi.kbnbfc.in/admin/user/verification/ad/grid	GET
https://dblapi.kbnbfc.in/admin/user/verification/pd/grid	GET
https://dblapi.kbnbfc.in/admin/user/verification/pd/riskchecklist	POST
https://dblapi.kbnbfc.in/admin/user/verification/pd/view	GET
https://dblapi.kbnbfc.in/admin/vastransaction	GET
https://dblapi.kbnbfc.in/admin/verification/agent	GET
https://dblapi.kbnbfc.in/admin/verification/agent	POST
https://dblapi.kbnbfc.in/admin/verification/agent	PUT
https://dblapi.kbnbfc.in/admin/verification/agent/list	GET
https://dblapi.kbnbfc.in/admin/verification/agent/priority	ANY
https://dblapi.kbnbfc.in/admin/verification/agent/reallocation	GET
https://dblapi.kbnbfc.in/admin/verification/agent/reallocation	POST
https://dblapi.kbnbfc.in/admin/verification/agent/reallocation/submit	POST
https://dblapi.kbnbfc.in/admin/verification/allocation/manual	GET
https://dblapi.kbnbfc.in/admin/verification/allocation/manual	POST
https://dblapi.kbnbfc.in/admin/verification/allocation/manual/history	GET
https://dblapi.kbnbfc.in/admin/verification/allocation/priority	GET
https://dblapi.kbnbfc.in/admin/verification/allocation/priority	POST
https://dblapi.kbnbfc.in/admin/verification/allocation/priority/history	GET
https://dblapi.kbnbfc.in/admin/verification/bucket	GET
https://dblapi.kbnbfc.in/admin/verification/bucket	POST
https://dblapi.kbnbfc.in/admin/verification/bucket	PUT
https://dblapi.kbnbfc.in/admin/verification/config/city	GET
https://dblapi.kbnbfc.in/admin/verification/config/city	POST
https://dblapi.kbnbfc.in/admin/verification/config/city/grid	GET
https://dblapi.kbnbfc.in/admin/verification/config/city/grid	POST
https://dblapi.kbnbfc.in/admin/verification/config/city/grid	PUT
https://dblapi.kbnbfc.in/admin/verification/config/city/history	GET
https://dblapi.kbnbfc.in/admin/verification/config/city/nach	POST
https://dblapi.kbnbfc.in/admin/verification/config/config/history	GET
https://dblapi.kbnbfc.in/admin/verification/config/config/history	POST
https://dblapi.kbnbfc.in/admin/verification/reallocation	GET
https://dblapi.kbnbfc.in/admin/verification/reallocation	POST
https://dblapi.kbnbfc.in/admin/verification/user/address	POST
https://dblapi.kbnbfc.in/admin/verification/user/appointment	POST
https://dblapi.kbnbfc.in/admin/verification/user/document	POST
https://dblapi.kbnbfc.in/admin/verification/user/document	PUT
https://dblapi.kbnbfc.in/admin/verification/user/status	POST
https://dblapi.kbnbfc.in/admin/vivriti	GET
https://dblapi.kbnbfc.in/admin/vivriti	POST
https://dblapi.kbnbfc.in/admin/vkyc/agent-mode-change	POST
https://dblapi.kbnbfc.in/admin/vkyc/agents	GET
https://dblapi.kbnbfc.in/admin/vkyc/agents/add	POST
https://dblapi.kbnbfc.in/admin/vkyc/agents/config	GET
https://dblapi.kbnbfc.in/admin/vkyc/agents/list	GET
https://dblapi.kbnbfc.in/admin/vkyc/agents/list	POST
https://dblapi.kbnbfc.in/admin/vkyc/auditvkyc	POST
https://dblapi.kbnbfc.in/admin/vkyc/callstatus	POST
https://dblapi.kbnbfc.in/admin/vkyc/config	GET
https://dblapi.kbnbfc.in/admin/vkyc/config	POST
https://dblapi.kbnbfc.in/admin/vkyc/currentActivity	GET
https://dblapi.kbnbfc.in/admin/vkyc/delete	POST
https://dblapi.kbnbfc.in/admin/vkyc/grid	GET
https://dblapi.kbnbfc.in/admin/vkyc/history	GET
https://dblapi.kbnbfc.in/admin/vkyc/login	GET
https://dblapi.kbnbfc.in/admin/vkyc/login	POST
https://dblapi.kbnbfc.in/admin/vkyc/logout	POST
https://dblapi.kbnbfc.in/admin/vkyc/markAgentOffline	POST
https://dblapi.kbnbfc.in/admin/vkyc/presignedurl	GET
https://dblapi.kbnbfc.in/admin/vkyc/processpdoffer	POST
https://dblapi.kbnbfc.in/admin/vkyc/register	POST
https://dblapi.kbnbfc.in/admin/vkyc/submit	POST
https://dblapi.kbnbfc.in/admin/vkyc/uploadvkycvideo	POST
https://dblapi.kbnbfc.in/admin/vkyc/userQueue	GET
https://dblapi.kbnbfc.in/admin/vkyc/userinfo	GET
https://dblapi.kbnbfc.in/admin/vkyc/webhook	POST
https://dblapi.kbnbfc.in/admin/whatsapp	GET
https://dblapi.kbnbfc.in/admin/whatsapp	POST
https://dblapi.kbnbfc.in/admin/whatsapp	PUT
https://dblapi.kbnbfc.in/admin/whatsapp/callback	GET
https://dblapi.kbnbfc.in/admin/whatsapp/callback	POST
https://dblapi.kbnbfc.in/admin/whatsapp/grid	GET
https://dblapi.kbnbfc.in/admin/whatsapp/grid	POST
https://dblapi.kbnbfc.in/admin/whatsapp/import	GET
https://dblapi.kbnbfc.in/admin/whatsapp/import	POST
https://dblapi.kbnbfc.in/admin/whatsapp/import/numbermanagement	POST
https://dblapi.kbnbfc.in/admin/whatsapp/notification	GET
https://dblapi.kbnbfc.in/admin/whatsapp/presignedurl	GET
https://dblapi.kbnbfc.in/admin/whatsapp/schedule	POST
https://dblapi.kbnbfc.in/admin/whatsapp/sns	POST
https://dblapi.kbnbfc.in/admin/whatsapp/templatechecks	GET
https://dblapi.kbnbfc.in/admin/whatsapp/templates	GET
https://dblapi.kbnbfc.in/dbl/aa	POST
https://dblapi.kbnbfc.in/dbl/aa/callback	POST
https://dblapi.kbnbfc.in/dbl/aa/rc	POST
https://dblapi.kbnbfc.in/dbl/aa/status	POST
https://dblapi.kbnbfc.in/dbl/auth/token	GET
https://dblapi.kbnbfc.in/dbl/banner	GET
https://dblapi.kbnbfc.in/dbl/banner	POST
https://dblapi.kbnbfc.in/dbl/banner	PUT
https://dblapi.kbnbfc.in/dbl/banner/grid	GET
https://dblapi.kbnbfc.in/dbl/category-main	GET
https://dblapi.kbnbfc.in/dbl/checkstatus	POST
https://dblapi.kbnbfc.in/dbl/contactus/freshchat/restoreid	GET
https://dblapi.kbnbfc.in/dbl/contactus/freshchat/restoreid	POST
https://dblapi.kbnbfc.in/dbl/contactus/query	POST
https://dblapi.kbnbfc.in/dbl/contactus/request	POST
https://dblapi.kbnbfc.in/dbl/contactus/rule	GET
https://dblapi.kbnbfc.in/dbl/contactus/unlockpdf	POST
https://dblapi.kbnbfc.in/dbl/customtoken	GET
https://dblapi.kbnbfc.in/dbl/dbl-bre-callback/10177	POST
https://dblapi.kbnbfc.in/dbl/digilocker/document-main	GET
https://dblapi.kbnbfc.in/dbl/digilocker/document-main	POST
https://dblapi.kbnbfc.in/dbl/digitalkyc/cafData	POST
https://dblapi.kbnbfc.in/dbl/digitalkyc/facematch	POST
https://dblapi.kbnbfc.in/dbl/digitalkyc/pan	GET
https://dblapi.kbnbfc.in/dbl/digitalkyc/status	GET
https://dblapi.kbnbfc.in/dbl/digitalkyc/verifiedaddress	POST
https://dblapi.kbnbfc.in/dbl/disbursal	POST
https://dblapi.kbnbfc.in/dbl/disbursal/pennycheck	POST
https://dblapi.kbnbfc.in/dbl/dlEvent	POST
https://dblapi.kbnbfc.in/dbl/document	GET
https://dblapi.kbnbfc.in/dbl/document	PUT
https://dblapi.kbnbfc.in/dbl/document/preview	GET
https://dblapi.kbnbfc.in/dbl/document/sign	POST
https://dblapi.kbnbfc.in/dbl/document/upload	POST
https://dblapi.kbnbfc.in/dbl/ebpp/fetch	POST
https://dblapi.kbnbfc.in/dbl/ebpp/payment	POST
https://dblapi.kbnbfc.in/dbl/ebpp/status	POST
https://dblapi.kbnbfc.in/dbl/edi/partner/loan/foreclosure	POST
https://dblapi.kbnbfc.in/dbl/email	ANY
https://dblapi.kbnbfc.in/dbl/email/pepipost	POST
https://dblapi.kbnbfc.in/dbl/email/send	POST
https://dblapi.kbnbfc.in/dbl/esign/esignDetails	GET
https://dblapi.kbnbfc.in/dbl/esign/esignRedirect	POST
https://dblapi.kbnbfc.in/dbl/esthenos	POST
https://dblapi.kbnbfc.in/dbl/faq	GET
https://dblapi.kbnbfc.in/dbl/faq/feedback	GET
https://dblapi.kbnbfc.in/dbl/faq/feedback	POST
https://dblapi.kbnbfc.in/dbl/fetchall	GET
https://dblapi.kbnbfc.in/dbl/generateOtp	PUT
https://dblapi.kbnbfc.in/dbl/geolocation	POST
https://dblapi.kbnbfc.in/dbl/getbalance	POST
https://dblapi.kbnbfc.in/dbl/home	ANY
https://dblapi.kbnbfc.in/dbl/incomeverify	GET
https://dblapi.kbnbfc.in/dbl/incomeverify/aa	POST
https://dblapi.kbnbfc.in/dbl/incomeverify/bank	GET
https://dblapi.kbnbfc.in/dbl/incomeverify/bank	POST
https://dblapi.kbnbfc.in/dbl/incomeverify/netbanking	POST
https://dblapi.kbnbfc.in/dbl/incomeverify/novel/cb	POST
https://dblapi.kbnbfc.in/dbl/incomeverify/statement	GET
https://dblapi.kbnbfc.in/dbl/incomeverify/statement	POST
https://dblapi.kbnbfc.in/dbl/lead/data/19092	GET
https://dblapi.kbnbfc.in/dbl/lead/token	POST
https://dblapi.kbnbfc.in/dbl/ledger/transactions	GET
https://dblapi.kbnbfc.in/dbl/legal	GET
https://dblapi.kbnbfc.in/dbl/loan/enach	GET
https://dblapi.kbnbfc.in/dbl/loan/enach	POST
https://dblapi.kbnbfc.in/dbl/loan/enach	PUT
https://dblapi.kbnbfc.in/dbl/loan/foreclosure/event	POST
https://dblapi.kbnbfc.in/dbl/loan/foreclosure/reasons	GET
https://dblapi.kbnbfc.in/dbl/loan/la	GET
https://dblapi.kbnbfc.in/dbl/loan/settlement	POST
https://dblapi.kbnbfc.in/dbl/loanDoc/shortLink	POST
https://dblapi.kbnbfc.in/dbl/loanPenalty/shortLink	POST
https://dblapi.kbnbfc.in/dbl/login	GET
https://dblapi.kbnbfc.in/dbl/login	POST
https://dblapi.kbnbfc.in/dbl/me	GET
https://dblapi.kbnbfc.in/dbl/me	PUT
https://dblapi.kbnbfc.in/dbl/me/aadhaarwarning	GET
https://dblapi.kbnbfc.in/dbl/me/additionalDetails	POST
https://dblapi.kbnbfc.in/dbl/me/additionalDetails/deviceInfo	POST
https://dblapi.kbnbfc.in/dbl/me/adprecheck	GET
https://dblapi.kbnbfc.in/dbl/me/algobeginprocessing	POST
https://dblapi.kbnbfc.in/dbl/me/algodata	GET
https://dblapi.kbnbfc.in/dbl/me/amazonscrapping	POST
https://dblapi.kbnbfc.in/dbl/me/amazonscrapping/otp	POST
https://dblapi.kbnbfc.in/dbl/me/apprating	PUT
https://dblapi.kbnbfc.in/dbl/me/appsflyer	POST
https://dblapi.kbnbfc.in/dbl/me/autodebit	GET
https://dblapi.kbnbfc.in/dbl/me/bank	GET
https://dblapi.kbnbfc.in/dbl/me/bank	PUT
https://dblapi.kbnbfc.in/dbl/me/checkaccexist	POST
https://dblapi.kbnbfc.in/dbl/me/ckyc/details	GET
https://dblapi.kbnbfc.in/dbl/me/ckyc/details	POST
https://dblapi.kbnbfc.in/dbl/me/ckyc/download	POST
https://dblapi.kbnbfc.in/dbl/me/ckyc/download	PUT
https://dblapi.kbnbfc.in/dbl/me/cscbrequest	GET
https://dblapi.kbnbfc.in/dbl/me/cscbrequest	PUT
https://dblapi.kbnbfc.in/dbl/me/delete	GET
https://dblapi.kbnbfc.in/dbl/me/digitap/nb	POST
https://dblapi.kbnbfc.in/dbl/me/dnd	GET
https://dblapi.kbnbfc.in/dbl/me/dnd	POST
https://dblapi.kbnbfc.in/dbl/me/doc	PUT
https://dblapi.kbnbfc.in/dbl/me/doc/kycoptions	GET
https://dblapi.kbnbfc.in/dbl/me/eligible	GET
https://dblapi.kbnbfc.in/dbl/me/email	GET
https://dblapi.kbnbfc.in/dbl/me/email	PUT
https://dblapi.kbnbfc.in/dbl/me/esign/cb	POST
https://dblapi.kbnbfc.in/dbl/me/ewaloan	PUT
https://dblapi.kbnbfc.in/dbl/me/flipkartscrapping	POST
https://dblapi.kbnbfc.in/dbl/me/geo	POST
https://dblapi.kbnbfc.in/dbl/me/kycLinkStatus	POST
https://dblapi.kbnbfc.in/dbl/me/lacd	PUT
https://dblapi.kbnbfc.in/dbl/me/livingindex	POST
https://dblapi.kbnbfc.in/dbl/me/loan	GET
https://dblapi.kbnbfc.in/dbl/me/loanhistory	GET
https://dblapi.kbnbfc.in/dbl/me/loans	GET
https://dblapi.kbnbfc.in/dbl/me/mfilimit	POST
https://dblapi.kbnbfc.in/dbl/me/nachmandate	GET
https://dblapi.kbnbfc.in/dbl/me/nachmandate	POST
https://dblapi.kbnbfc.in/dbl/me/nachmandate	PUT
https://dblapi.kbnbfc.in/dbl/me/nachmandate/cb/37144	POST
https://dblapi.kbnbfc.in/dbl/me/nachmandate/status	GET
https://dblapi.kbnbfc.in/dbl/me/offers	GET
https://dblapi.kbnbfc.in/dbl/me/otp	GET
https://dblapi.kbnbfc.in/dbl/me/otp	POST
https://dblapi.kbnbfc.in/dbl/me/otp	PUT
https://dblapi.kbnbfc.in/dbl/me/partner/incomeverifydoc	POST
https://dblapi.kbnbfc.in/dbl/me/partner/otp	POST
https://dblapi.kbnbfc.in/dbl/me/partnerotp	POST
https://dblapi.kbnbfc.in/dbl/me/payment	GET
https://dblapi.kbnbfc.in/dbl/me/payment/ad	GET
https://dblapi.kbnbfc.in/dbl/me/payment/addpayments	POST
https://dblapi.kbnbfc.in/dbl/me/payment/cb/21855	GET
https://dblapi.kbnbfc.in/dbl/me/payment/cb/29332	POST
https://dblapi.kbnbfc.in/dbl/me/payment/convFee	GET
https://dblapi.kbnbfc.in/dbl/me/payment/extradata	POST
https://dblapi.kbnbfc.in/dbl/me/payment/history	GET
https://dblapi.kbnbfc.in/dbl/me/payment/issue	POST
https://dblapi.kbnbfc.in/dbl/me/payment/order	GET
https://dblapi.kbnbfc.in/dbl/me/payment/order	POST
https://dblapi.kbnbfc.in/dbl/me/payment/undopayments	POST
https://dblapi.kbnbfc.in/dbl/me/profile/doc/aadhaar	POST
https://dblapi.kbnbfc.in/dbl/me/profile/vkyc	GET
https://dblapi.kbnbfc.in/dbl/me/profile/vkyc	POST
https://dblapi.kbnbfc.in/dbl/me/profile/vpd	GET
https://dblapi.kbnbfc.in/dbl/me/rescheduleloan	GET
https://dblapi.kbnbfc.in/dbl/me/rescheduleloan	PUT
https://dblapi.kbnbfc.in/dbl/me/rescheduleloan/tnc	GET
https://dblapi.kbnbfc.in/dbl/me/rescheduleloan/tnc	POST
https://dblapi.kbnbfc.in/dbl/me/rescheduleloan/updateduedate	POST
https://dblapi.kbnbfc.in/dbl/me/signout	GET
https://dblapi.kbnbfc.in/dbl/me/sync	POST
https://dblapi.kbnbfc.in/dbl/me/tnc	GET
https://dblapi.kbnbfc.in/dbl/me/tnc	POST
https://dblapi.kbnbfc.in/dbl/me/tnc/11818	GET
https://dblapi.kbnbfc.in/dbl/me/tnc/31444	POST
https://dblapi.kbnbfc.in/dbl/me/webpayment	POST
https://dblapi.kbnbfc.in/dbl/me/webpayment/cb/21951	GET
https://dblapi.kbnbfc.in/dbl/me/webpayment/cb/12484	POST
https://dblapi.kbnbfc.in/dbl/mfilterit	POST
https://dblapi.kbnbfc.in/dbl/misc/6e42e9b6c48d65b07763484ba19d19cb	GET
https://dblapi.kbnbfc.in/dbl/misc/bank	GET
https://dblapi.kbnbfc.in/dbl/misc/companies	GET
https://dblapi.kbnbfc.in/dbl/misc/companies	PUT
https://dblapi.kbnbfc.in/dbl/misc/perfios/banks	GET
https://dblapi.kbnbfc.in/dbl/misc/pincode	GET
https://dblapi.kbnbfc.in/dbl/misc/pincode/emp	GET
https://dblapi.kbnbfc.in/dbl/myaccount	GET
https://dblapi.kbnbfc.in/dbl/nachmandate/cb/25108	POST
https://dblapi.kbnbfc.in/dbl/nachmandate/mandatemodes	GET
https://dblapi.kbnbfc.in/dbl/paymentlink/amount	POST
https://dblapi.kbnbfc.in/dbl/paymentlink/generate	POST
https://dblapi.kbnbfc.in/dbl/paymentlink/loandetails	GET
https://dblapi.kbnbfc.in/dbl/paymentlink/order/create	POST
https://dblapi.kbnbfc.in/dbl/paymentlink/order/status	POST
https://dblapi.kbnbfc.in/dbl/paymentlink/paymentdetails	POST
https://dblapi.kbnbfc.in/dbl/paymentlink/pggateways	GET
https://dblapi.kbnbfc.in/dbl/paymentlink/visit	POST
https://dblapi.kbnbfc.in/dbl/paymentlink/visit/convFee	GET
https://dblapi.kbnbfc.in/dbl/pdf	PUT
https://dblapi.kbnbfc.in/dbl/products	GET
https://dblapi.kbnbfc.in/dbl/products/calcfees	GET
https://dblapi.kbnbfc.in/dbl/products/info	GET
https://dblapi.kbnbfc.in/dbl/repaymentlink	GET
https://dblapi.kbnbfc.in/dbl/repaymentlink	POST
https://dblapi.kbnbfc.in/dbl/repaymentlink/cb/10385	GET
https://dblapi.kbnbfc.in/dbl/repaymentlink/cb/26432	POST
https://dblapi.kbnbfc.in/dbl/sms/deliverystatus/33914	GET
https://dblapi.kbnbfc.in/dbl/sms/dlstatus/airtel	POST
https://dblapi.kbnbfc.in/dbl/sms/dlstatus/17444	GET
https://dblapi.kbnbfc.in/dbl/sms/getstatus	GET
https://dblapi.kbnbfc.in/dbl/sms/send	POST
https://dblapi.kbnbfc.in/dbl/sms/status	POST
https://dblapi.kbnbfc.in/dbl/st	GET
https://dblapi.kbnbfc.in/dbl/transaction/collectionPartnerUpdateInwards/creditmate	POST
https://dblapi.kbnbfc.in/dbl/transaction/lendingPartnerUpdateInwards/mipay	POST
https://dblapi.kbnbfc.in/dbl/vaspayment	POST
https://dblapi.kbnbfc.in/dbl/vkyc/agents/notification/response	POST
https://dblapi.kbnbfc.in/dbl/vkyc/refId/35155	POST
https://dblapi.kbnbfc.in/dbl/voice/deliverystatus/39614	POST
https://dblapi.kbnbfc.in/dbl/voice/dlstatus/27844	POST
https://dblapi.kbnbfc.in/dbl/voice/send	POST
https://dblapi.kbnbfc.in/dbl/voice/status	POST
https://dblapi.kbnbfc.in/dbl/vpd/agents/notification/response	POST
https://dblapi.kbnbfc.in/dbl/webhook-main/32370	GET
https://dblapi.kbnbfc.in/dbl/webhook-main/11367	POST
https://dblapi.kbnbfc.in/dbl/webhook-sb/26705	POST
https://dblapi.kbnbfc.in/dbl/webhook/28124	POST
https://dblapi.kbnbfc.in/dbldsa/get-otp	PUT
https://dblapi.kbnbfc.in/dbldsa/submit-otp	POST
https://dblapi.kbnbfc.in/edi/admin/edi-callback/exotel	POST
https://dblapi.kbnbfc.in/edi/edi-re-callback/12012	POST
https://dblapi.kbnbfc.in/popsadsa/dblkb-lambda-popsa-dsa	ANY
https://dblapi.kbnbfc.in/popsadsa/get-otp	PUT
https://dblapi.kbnbfc.in/popsadsa/submit-otp	POST
https://dblapi.kbnbfc.in/smssvc/sms/deliverystatus/42506	GET
https://dblapi.kbnbfc.in/smssvc/sms/dlstatus/38692	GET
https://j3yiqg5r7f.execute-api.ap-south-2.amazonaws.com/e	GET
https://j3yiqg5r7f.execute-api.ap-south-2.amazonaws.com/e/pets	GET
https://j3yiqg5r7f.execute-api.ap-south-2.amazonaws.com/e/pets	POST
https://j3yiqg5r7f.execute-api.ap-south-2.amazonaws.com/e/pets/39808	GET
`

var anyMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}

func parseEndpoints(raw string) []Endpoint {
	lines := strings.Split(raw, "\n")
	out := make([]Endpoint, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.EqualFold(line, "Url\tMethod") || strings.EqualFold(line, "Url Method") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		url := fields[0]
		method := strings.ToUpper(fields[len(fields)-1])
		out = append(out, Endpoint{URL: url, Method: method})
	}

	return out
}

func isWriteMethod(m string) bool {
	switch m {
	case "POST", "PUT", "PATCH", "DELETE":
		return true
	default:
		return false
	}
}

func expandTargets(endpoints []Endpoint) []Target {
	var out []Target
	for _, ep := range endpoints {
		if strings.EqualFold(ep.Method, "ANY") {
			for _, m := range anyMethods {
				out = append(out, Target{URL: ep.URL, Method: m, ExpandedFromAny: true})
			}
		} else {
			out = append(out, Target{URL: ep.URL, Method: strings.ToUpper(ep.Method), ExpandedFromAny: false})
		}
	}
	return out
}

func headersToMap(h http.Header) map[string]string {
	m := make(map[string]string, len(h))
	for k, v := range h {
		m[k] = strings.Join(v, ", ")
	}
	return m
}

func runOne(client *http.Client, t Target, allowWrite bool, bodyLimit int) Result {
	res := Result{
		URL:             t.URL,
		Method:          t.Method,
		TimestampUTC:    time.Now().UTC().Format(time.RFC3339Nano),
		ExpandedFromAny: t.ExpandedFromAny,
	}

	if isWriteMethod(t.Method) && !allowWrite {
		res.Skipped = true
		res.Reason = "write method skipped (use -allow-write to enable)"
		return res
	}

	var body io.Reader
	if t.Method == "POST" || t.Method == "PUT" || t.Method == "PATCH" {
		body = bytes.NewBufferString(`{}`)
	}

	req, err := http.NewRequest(t.Method, t.URL, body)
	if err != nil {
		res.Error = "RequestBuildError"
		res.ErrorMessage = err.Error()
		return res
	}

	req.Header.Set("User-Agent", "noauth-api-probe/1.0")
	req.Header.Set("Accept", "*/*")
	if t.Method == "POST" || t.Method == "PUT" || t.Method == "PATCH" {
		req.Header.Set("Content-Type", "application/json")
	}
	// Intentionally no Authorization header.

	start := time.Now()
	resp, err := client.Do(req)
	res.ElapsedMs = float64(time.Since(start).Milliseconds())

	if err != nil {
		res.Error = "RequestError"
		res.ErrorMessage = err.Error()
		return res
	}
	defer resp.Body.Close()

	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		res.Error = "ReadBodyError"
		res.ErrorMessage = readErr.Error()
		res.StatusCode = resp.StatusCode
		res.ResponseHeaders = headersToMap(resp.Header)
		return res
	}

	bodyText := string(raw)
	truncated := false
	if len(bodyText) > bodyLimit {
		bodyText = bodyText[:bodyLimit]
		truncated = true
	}

	res.StatusCode = resp.StatusCode
	res.ResponseHeaders = headersToMap(resp.Header)
	res.ResponseBody = bodyText
	res.ResponseBodyTruncated = truncated
	res.ResponseSizeBytes = len(raw)
	if resp.StatusCode >= 400 {
		res.HTTPError = true
	}

	return res
}

func main() {
	timeoutSec := flag.Int("timeout", 20, "Per-request timeout in seconds")
	delayMs := flag.Int("delay-ms", 250, "Delay between requests in milliseconds")
	bodyLimit := flag.Int("body-limit", 5000, "Max response body characters to store")
	allowWrite := flag.Bool("allow-write", false, "Enable POST/PUT/PATCH/DELETE calls")
	out := flag.String("out", fmt.Sprintf("noauth_probe_results_%d.json", time.Now().Unix()), "Output JSON file")
	flag.Parse()

	endpoints := parseEndpoints(endpointTSV)
	targets := expandTargets(endpoints)
	results := make([]Result, 0, len(targets))

	client := &http.Client{
		Timeout: time.Duration(*timeoutSec) * time.Second,
	}

	fmt.Printf("Base endpoints loaded: %d\n", len(endpoints))
	fmt.Printf("Total requests to attempt: %d\n", len(targets))
	fmt.Printf("Write methods enabled: %v\n", *allowWrite)
	fmt.Println("Starting...")

	for i, t := range targets {
		fmt.Printf("[%d/%d] %s %s\n", i+1, len(targets), t.Method, t.URL)
		r := runOne(client, t, *allowWrite, *bodyLimit)
		results = append(results, r)
		time.Sleep(time.Duration(*delayMs) * time.Millisecond)
	}

	f, err := os.Create(*out)
	if err != nil {
		fmt.Printf("Failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(results); err != nil {
		fmt.Printf("Failed to write output JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nDone. Saved detailed output to: %s\n", *out)

	// Summary
	counts := map[string]int{}
	for _, r := range results {
		key := "ERROR"
		if r.Skipped {
			key = "SKIPPED"
		} else if r.StatusCode > 0 {
			key = fmt.Sprintf("%d", r.StatusCode)
		}
		counts[key]++
	}

	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Println("Summary:")
	for _, k := range keys {
		fmt.Printf("  %s: %d\n", k, counts[k])
	}
}

