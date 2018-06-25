package audittypes

import (
	"io"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"
)

const saMsgXML string = `
<?xml version="1.0" encoding="UTF-8"?>
<GovTalkMessage xmlns="http://www.govtalk.gov.uk/CM/envelope">
	<EnvelopeVersion>2.0</EnvelopeVersion>
	<Header>
		<MessageDetails>
			<Class>HMRC-SA-SA800-ATT</Class>
			<Qualifier>request</Qualifier>
			<Function>submit</Function>
			<TransactionID/>
			<CorrelationID>CF53E1122A4347A5A1289B33659EC1AA</CorrelationID>
			<ResponseEndPoint>IR-SERVICE-ENDPOINT-EXTRA-4</ResponseEndPoint>
			<Transformation>XML</Transformation>
			<GatewayTimestamp>2016-03-11T10:49:00.763</GatewayTimestamp>
		</MessageDetails>
		<SenderDetails>
			<IDAuthentication>
				<SenderID>0000000111111111</SenderID>
				<Authentication>
					<Method>clear</Method>
					<Role>principal</Role>
					<Value>**********</Value>
				</Authentication>
			</IDAuthentication>
			<X509Certificate/>
		</SenderDetails>
	</Header>
	<GovTalkDetails>
		<Keys>
			<Key Type='UTR'>5566778899</Key>
		</Keys>
		<TargetDetails>
			<Organisation>IR</Organisation>
		</TargetDetails>
		<ChannelRouting>
			<Channel>
				<URI>1234</URI>
				<Product>X-Meta</Product>
				<Version>2.02</Version>
			</Channel>
		</ChannelRouting>
		<GatewayAdditions>
			<Submitter xmlns="http://www.govtalk.gov.uk/gateway/submitterdetails">
				<AgentDetails/>
				<SubmitterDetails>
					<RegistrationCategory>Organisation</RegistrationCategory>
					<UserType>Principal</UserType>
					<CredentialRole>User</CredentialRole>
					<CredentialIdentifier>0000000112233445</CredentialIdentifier>
				</SubmitterDetails>
			</Submitter>
		</GatewayAdditions>
	</GovTalkDetails>
	<Body>
		<IRenvelope xmlns="http://www.govtalk.gov.uk/taxation/SA/SA800/10-11/1">
			<IRheader>
				<Keys>
					<Key Type='UTR'>5566778899</Key>
				</Keys>
				<PeriodEnd>2011-04-05</PeriodEnd>
				<IRmark Type='generic'>H6Xs97sZcIrPURNCwDIteNKw5qo=</IRmark>
				<Sender>Partnership</Sender>
			</IRheader>
			<SApartnership>
				<PartnershipName>ABCDEFGHIJKLMNOPQRSTUVWXYZ123456</PartnershipName>
				<Partnership>
					<BusinessInvestmentIncome>
						<LandProperty>yes</LandProperty>
						<Foreign>yes</Foreign>
						<Trade>yes</Trade>
						<ChargeableAssets>yes</ChargeableAssets>
					</BusinessInvestmentIncome>
					<TradingAndProfessionalIncomes>
						<Income>
							<NameOfBusiness>Horsey Things</NameOfBusiness>
							<Description>Equestrian supplies</Description>
							<AccountingPeriodStart>2010-01-01</AccountingPeriodStart>
							<AccountingPeriodEnd>2010-12-31</AccountingPeriodEnd>
							<DateOfCommencement>2008-12-31</DateOfCommencement>
							<CapitalAllowancesSummary>
								<Cars>1576.00</Cars>
								<PlantAndMachinery>4290.00</PlantAndMachinery>
								<TotalAllowances>5866.00</TotalAllowances>
								<TotalBalancingCharges>0.00</TotalBalancingCharges>
							</CapitalAllowancesSummary>
							<AdditionalInformation>
								<Line>Az9&amp;&apos;(),-&apos;:.*^!+_@{}</Line>
							</AdditionalInformation>
							<IncomeAndExpenses>
								<ExcludeVAT>yes</ExcludeVAT>
								<SalesBusinessIncome>56300.00</SalesBusinessIncome>
								<DisallowableDepreciation>1232.00</DisallowableDepreciation>
								<CostOfSales>39520.00</CostOfSales>
								<SubcontractorCosts>4121.25</SubcontractorCosts>
								<OtherDirectCosts>14782.21</OtherDirectCosts>
								<GrossProfitLoss>-2123.46</GrossProfitLoss>
								<OtherIncomeProfits>500.00</OtherIncomeProfits>
								<EmployeeCosts>9800.00</EmployeeCosts>
								<PremisesCosts>3776.00</PremisesCosts>
								<Repairs>300.00</Repairs>
								<GeneralAdministrativeExpenses>915.00</GeneralAdministrativeExpenses>
								<MotorExpenses>1040.00</MotorExpenses>
								<LegalandProfessionalCosts>225.00</LegalandProfessionalCosts>
								<OtherFinanceCharges>273.00</OtherFinanceCharges>
								<DepreciationAndLoss>1232.00</DepreciationAndLoss>
								<OtherExpenses>444.00</OtherExpenses>
								<TotalExpenses>18005.00</TotalExpenses>
								<NetProfitOrLoss>-19628.46</NetProfitOrLoss>
							</IncomeAndExpenses>
							<TaxAdjustments>
								<AdditionsToNetProfit>
									<DisallowableExpenses>1232.00</DisallowableExpenses>
									<BalancingCharge>0.00</BalancingCharge>
									<TotalAdditionsToNetProfit>1232.00</TotalAdditionsToNetProfit>
								</AdditionsToNetProfit>
								<DeductionsFromNetProfit>
									<CapitalAllowances>5866.00</CapitalAllowances>
									<TotalDeductionsFromNetProfit>5866.00</TotalDeductionsFromNetProfit>
								</DeductionsFromNetProfit>
								<NetProfitLoss>-24262.46</NetProfitLoss>
							</TaxAdjustments>
							<AdjustmentsForTaxableProfitOrLoss>
								<NetProfit>0.00</NetProfit>
								<AllowableLoss>5359.00</AllowableLoss>
								<ProvisionalProfitLoss>yes</ProvisionalProfitLoss>
							</AdjustmentsForTaxableProfitOrLoss>
							<SummaryOfBalanceSheet>
								<Assets>
									<PlantMachineryAndMotorVehicles>15670.00</PlantMachineryAndMotorVehicles>
									<OtherFixedAssets>4238.00</OtherFixedAssets>
									<StockAndWorkInProgress>32399.00</StockAndWorkInProgress>
									<DebtorsPrepaymentsEtc>600.00</DebtorsPrepaymentsEtc>
									<BankEtcBalances>15115.00</BankEtcBalances>
									<CashInHand>1500.00</CashInHand>
									<TotalAssets>69522.00</TotalAssets>
								</Assets>
								<Liabilities>
									<LoansAndOverdrawnBankAccounts>9700.00</LoansAndOverdrawnBankAccounts>
									<OtherLiabilities>14567890.12</OtherLiabilities>
									<TotalLiabilities>14577590.12</TotalLiabilities>
								</Liabilities>
								<NetBusinessAssets>-14508068.12</NetBusinessAssets>
								<PartnersCurrentAndCapitalAccounts>
									<BalanceAtStartOfPeriod>50366.00</BalanceAtStartOfPeriod>
									<NetProfitLoss>-725.00</NetProfitLoss>
									<CapitalIntroduced>11400.00</CapitalIntroduced>
									<Drawings>124567890.12</Drawings>
									<BalanceAtEndOfPeriod>-124506849.12</BalanceAtEndOfPeriod>
								</PartnersCurrentAndCapitalAccounts>
							</SummaryOfBalanceSheet>
						</Income>
					</TradingAndProfessionalIncomes>
					<PartnershipStatement>
						<PartnershipInformation>
							<PeriodStart>2010-01-01</PeriodStart>
							<PeriodEnd>2010-12-31</PeriodEnd>
							<NatureOfTrade>Equestrian supplies</NatureOfTrade>
							<TradeProfit>0.00</TradeProfit>
							<TradeLoss>5359.00</TradeLoss>
							<UntaxedUKsavingsIncome>1245.00</UntaxedUKsavingsIncome>
							<UntaxedForeignSavingsIncome>800.00</UntaxedForeignSavingsIncome>
							<ForeignDividends>2666.67</ForeignDividends>
							<OtherUntaxedForeignIncome>13435.00</OtherUntaxedForeignIncome>
							<ProfitLossOnUKlandAndProperty>17120.00</ProfitLossOnUKlandAndProperty>
							<TaxedSavingsAtLowerDividendRate>750.00</TaxedSavingsAtLowerDividendRate>
							<TaxedSavings>1665.00</TaxedSavings>
							<UKincomeTax>333.00</UKincomeTax>
							<UKnotionalIncomeTax>341.67</UKnotionalIncomeTax>
							<ForeignTaxPaid>1105.00</ForeignTaxPaid>
							<TotalDisposalChargeableAssets>20900.00</TotalDisposalChargeableAssets>
						</PartnershipInformation>
						<PartnerDetails>
							<PartnerName>Mary Anybody</PartnerName>
							<PartnerAddress>
								<Line>16 Anywhere Lane</Line>
								<Line>Anywhere</Line>
							</PartnerAddress>
							<PartnerUTR>8777777771</PartnerUTR>
							<TradeLoss>2680.00</TradeLoss>
							<UKsavingsIncome>623.00</UKsavingsIncome>
							<UntaxedForeignSavingsIncome>400.00</UntaxedForeignSavingsIncome>
							<OverseasDividends>1666.67</OverseasDividends>
							<OtherUntaxedForeignIncome>7035.00</OtherUntaxedForeignIncome>
							<ProfitLossOnUKlandAndProperty>8560.00</ProfitLossOnUKlandAndProperty>
							<TaxedSavingsAtLowerDividendRate>375.00</TaxedSavingsAtLowerDividendRate>
							<TaxedSavings>833.00</TaxedSavings>
							<UKincomeTax>167.00</UKincomeTax>
							<UKnotionalIncomeTax>304.67</UKnotionalIncomeTax>
							<ForeignTaxPaid>553.00</ForeignTaxPaid>
							<DisposalOfChargeableAssets>20000.00</DisposalOfChargeableAssets>
						</PartnerDetails>
						<PartnerDetails>
							<PartnerName>Ivor Anybody</PartnerName>
							<PartnerAddress>
								<Line>16 Anywhere Lane</Line>
								<Line>New Estate</Line>
								<Line>Off Nowhere Boulevard</Line>
								<ShortLine>Anywhere</ShortLine>
								<PostCode>XX11 1XX</PostCode>
							</PartnerAddress>
							<PartnerUTR>4222222221</PartnerUTR>
							<TradeLoss>2679.00</TradeLoss>
							<UKsavingsIncome>622.00</UKsavingsIncome>
							<UntaxedForeignSavingsIncome>400.00</UntaxedForeignSavingsIncome>
							<OverseasDividends>1000.00</OverseasDividends>
							<OtherUntaxedForeignIncome>6400.00</OtherUntaxedForeignIncome>
							<ProfitLossOnUKlandAndProperty>8560.00</ProfitLossOnUKlandAndProperty>
							<TaxedSavingsAtLowerDividendRate>375.00</TaxedSavingsAtLowerDividendRate>
							<TaxedSavings>832.00</TaxedSavings>
							<UKincomeTax>166.00</UKincomeTax>
							<UKnotionalIncomeTax>37.00</UKnotionalIncomeTax>
							<ForeignTaxPaid>552.00</ForeignTaxPaid>
							<DisposalOfChargeableAssets>900.00</DisposalOfChargeableAssets>
						</PartnerDetails>
					</PartnershipStatement>
					<OtherInformation>
						<IncomeNotIncludedElsewhere>yes</IncomeNotIncludedElsewhere>
						<DetailsIncorrect>yes</DetailsIncorrect>
						<TaxPayerAddress>
							<Line>The Quaintest Little Cottage</Line>
							<Line>423 Daintiest of Dainty Lane</Line>
							<Line>Yeardley Gobion in Towcester</Line>
							<ShortLine>Northamptonleshire</ShortLine>
							<PostCode>jklm opq</PostCode>
						</TaxPayerAddress>
						<AgentTelephone>01444 123456</AgentTelephone>
						<AgentName>Anyone Co</AgentName>
						<AgentAddress>
							<Line>High Street</Line>
							<Line>Anywhere</Line>
							<PostCode>jklm opq</PostCode>
						</AgentAddress>
						<IncludesProvisionalFigures>yes</IncludesProvisionalFigures>
						<NumberOfPartners>2</NumberOfPartners>
						<NominatedPartner>Mrs Mary Anybody</NominatedPartner>
						<FullPartnershipStatement>yes</FullPartnershipStatement>
						<PartnershipLandProperty>yes</PartnershipLandProperty>
						<PartnershipForeign>yes</PartnershipForeign>
						<PartnershipDisposalOfChargeableAssets>yes</PartnershipDisposalOfChargeableAssets>
						<PartnershipSavings>yes</PartnershipSavings>
					</OtherInformation>
				</Partnership>
				<PartnershipLandProperty>
					<PartnershipDetails>
						<ReturnPeriodStart>2010-01-01</ReturnPeriodStart>
						<ReturnPeriodEnd>2010-12-31</ReturnPeriodEnd>
					</PartnershipDetails>
					<FurnishedHolidayLettings>
						<Profit>0.00</Profit>
					</FurnishedHolidayLettings>
					<OtherProperty>
						<Income>
							<FurnishedHolidayLettingsProfits>0.00</FurnishedHolidayLettingsProfits>
							<RentsEtc>
								<Income>19997.00</Income>
							</RentsEtc>
							<TotalIncome>19997.00</TotalIncome>
						</Income>
						<Expenses>
							<RentEtc>1500.00</RentEtc>
							<RepairsEtc>742.00</RepairsEtc>
							<OtherExpenses>635.00</OtherExpenses>
							<TotalExpenses>2877.00</TotalExpenses>
						</Expenses>
						<NetProfitLoss>17120.00</NetProfitLoss>
						<ProfitLossForReturnPeriod>17120.00</ProfitLossForReturnPeriod>
					</OtherProperty>
				</PartnershipLandProperty>
				<PartnershipForeign>
					<ForeignIncomeSavingsLandProperty>
						<ReturnPeriodStart>2010-01-01</ReturnPeriodStart>
						<ReturnPeriodEnd>2010-12-31</ReturnPeriodEnd>
						<InterestEtc>
							<InterestEtcRow>
								<Country>Switzerland</Country>
								<Unremittable>no</Unremittable>
								<Gross>800.00</Gross>
								<ForeignTax>80.00</ForeignTax>
							</InterestEtcRow>
						</InterestEtc>
						<TotalIncomeRemittable>800.00</TotalIncomeRemittable>
						<Dividends>
							<DividendsRow>
								<Country>Denmark</Country>
								<Gross>400.00</Gross>
								<ForeignTax>40.00</ForeignTax>
							</DividendsRow>
							<DividendsRow>
								<Country>Italy</Country>
								<Gross>800.00</Gross>
								<ForeignTax>80.00</ForeignTax>
							</DividendsRow>
							<DividendsRow>
								<Country>Holland</Country>
								<Gross>1200.00</Gross>
								<ForeignTax>120.00</ForeignTax>
							</DividendsRow>
						</Dividends>
						<TotalDividendsIncomeRemittable>2400.00</TotalDividendsIncomeRemittable>
						<LandProperty>
							<Gross>635.00</Gross>
						</LandProperty>
						<ChargeablePremiums>
							<Country>USA</Country>
							<Gross>12800.00</Gross>
							<ForeignTax>825.00</ForeignTax>
						</ChargeablePremiums>
						<TotalGross>13435.00</TotalGross>
						<TotalForeignTax>1105.00</TotalForeignTax>
					</ForeignIncomeSavingsLandProperty>
					<IncomeLandPropertyAbroad>
						<IncomeLandAndPropertyAbroad>
							<Address>
								<Line>The Horse Ranch</Line>
								<Line>Houston Plains</Line>
								<Line>Texas</Line>
								<ShortLine>USA</ShortLine>
								<PostCode>jklm opq</PostCode>
							</Address>
							<TotalRentsEtc>20500.00</TotalRentsEtc>
							<Expenses>
								<RentRatesEtc>3020.00</RentRatesEtc>
								<RepairsAndRenewals>3770.00</RepairsAndRenewals>
								<LegalAndProfessionalCosts>600.00</LegalAndProfessionalCosts>
								<CostOfServices>825.00</CostOfServices>
								<OtherExpenses>590.00</OtherExpenses>
								<TotalExpenditure>8805.00</TotalExpenditure>
							</Expenses>
							<NetProfitLoss>11695.00</NetProfitLoss>
							<TaxAdjustments>
								<AdditionsToNetProfit>
									<PrivateUse>3020.00</PrivateUse>
									<TotalAdditions>3020.00</TotalAdditions>
								</AdditionsToNetProfit>
								<DeductionsFromNetProfit>
									<CapitalAllowances>1720.00</CapitalAllowances>
									<EnhancedCapitalAllowances>yes</EnhancedCapitalAllowances>
									<WearAndTear>195.00</WearAndTear>
									<TotalDeductions>1915.00</TotalDeductions>
								</DeductionsFromNetProfit>
							</TaxAdjustments>
							<AdjustedProfit>12800.00</AdjustedProfit>
						</IncomeLandAndPropertyAbroad>
						<TotalAllowableLoss>12.02</TotalAllowableLoss>
					</IncomeLandPropertyAbroad>
					<AdditionalInformation>
						<Line>123456789</Line>
					</AdditionalInformation>
				</PartnershipForeign>
				<PartnershipChargeableAssets>
					<Disposals>
						<DisposalRow>
							<Description>House Boat</Description>
							<Proceeds>20900.00</Proceeds>
							<FurtherInformation>123456789</FurtherInformation>
						</DisposalRow>
						<DisposalRow>
							<Unquoted>yes</Unquoted>
						</DisposalRow>
					</Disposals>
					<TotalProceeds>20900.00</TotalProceeds>
					<AdditionalInformation>
						<Line>123456789</Line>
					</AdditionalInformation>
				</PartnershipChargeableAssets>
				<PartnershipSavings>
					<Interest>
						<TaxNotDeducted>
							<PeriodStart>2010-01-01</PeriodStart>
							<PeriodEnd>2010-12-31</PeriodEnd>
							<UKbanksEtcGross>957.00</UKbanksEtcGross>
							<NationalSavings>288.00</NationalSavings>
							<TotalInterestTaxNotDeducted>1245.00</TotalInterestTaxNotDeducted>
						</TaxNotDeducted>
						<TaxDeducted>
							<UKbanksEtcNet>
								<Net>995.00</Net>
								<Tax>248.75</Tax>
								<Gross>1243.75</Gross>
							</UKbanksEtcNet>
							<OtherIncome>
								<Net>337.00</Net>
								<Tax>84.25</Tax>
								<Gross>421.25</Gross>
							</OtherIncome>
							<TotalTaxDeducted>333.00</TotalTaxDeducted>
							<TotalGross>1665.00</TotalGross>
						</TaxDeducted>
					</Interest>
					<Dividends>
						<ScripDividends>
							<Net>675.00</Net>
							<Tax>75.00</Tax>
							<Gross>750.00</Gross>
						</ScripDividends>
						<TotalNotionalTax>75.00</TotalNotionalTax>
						<TotalGross>750.00</TotalGross>
					</Dividends>
					<OtherIncome>
						<WithTaxDeducted>
							<Net>0.00</Net>
							<Tax>0.00</Tax>
							<Gross>0.00</Gross>
						</WithTaxDeducted>
					</OtherIncome>
				</PartnershipSavings>
				<Declaration>
					<PartnershipDeclaration>yes</PartnershipDeclaration>
				</Declaration>
				<MTR>
					<SA100>
						<YourPersonalDetails>
							<NationalInsuranceNumber>XX55443322</NationalInsuranceNumber>
						</YourPersonalDetails>
					</SA100>
				</MTR>
				<AttachedFiles>
					<Attachment FileFormat='zzz' Filename='blah.zzz' Description='some attachment'>
					</Attachment>
				</AttachedFiles>
			</SApartnership>
		</IRenvelope>
	</Body>
</GovTalkMessage>
`

func BenchmarkAudit20KbAttach(b *testing.B) {
	benchWithPayload(b, 20)
}

func BenchmarkAudit100KbAttach(b *testing.B) {
	benchWithPayload(b, 100)
}

func BenchmarkAudit250KbAttach(b *testing.B) {
	benchWithPayload(b, 250)
}

func BenchmarkAudit500KbAttach(b *testing.B) {
	benchWithPayload(b, 500)
}

func BenchmarkAudit1MegAttach(b *testing.B) {
	benchWithPayload(b, 1*1024)
}

func BenchmarkAudit20MegAttach(b *testing.B) {
	benchWithPayload(b, 20*1024)
}

func BenchmarkAudit50MegAttach(b *testing.B) {
	benchWithPayload(b, 50*1024)
}

func BenchmarkAudit100MegAttach(b *testing.B) {
	benchWithPayload(b, 100*1024)
}

func benchWithPayload(b *testing.B, payloadKbs int) {
	req := httptest.NewRequest("POST", "/some/rate/url?qq=zz", messageWithPayload(payloadKbs))

	for i := 0; i < b.N; i++ {
		event := &RATEAuditEvent{}
		spec := &AuditSpecification{}
		event.AppendRequest(req, spec)
	}
}

func messageWithPayload(payloadKbs int) io.ReadCloser {
	xml := saMsgXML
	if i := strings.Index(xml, "</Attachment>"); i != -1 {
		lenSlice := payloadKbs * 1024
		dat := make([]byte, lenSlice, lenSlice)
		newXML := xml[:i] + string(dat) + xml[i:]
		return ioutil.NopCloser(strings.NewReader(newXML))
	}
	panic("Didn't find Attachment element")
}
