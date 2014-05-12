/*
  btcrobot is a Bitcoin, Litecoin and Altcoin trading bot written in golang,
  it features multiple trading methods using technical analysis.

  Disclaimer:

  USE AT YOUR OWN RISK!

  The author of this project is NOT responsible for any damage or loss caused
  by this software. There can be bugs and the bot may not Tick as expected
  or specified. Please consider testing it first with paper trading /
  backtesting on historical data. Also look at the code to see what how
  it's working.

  Weibo:http://weibo.com/bocaicfa
*/

package strategy

import (
	. "common"
	. "config"
	"email"
	"logger"
	"strconv"
)

type HLCrossStrategy struct {
	PrevClosePrice float64
	PrevHighPrice  float64
	PrevLowPrice   float64
}

func init() {
	HLCross := new(HLCrossStrategy)
	Register("HLCross", HLCross)
}

//HLCross strategy
func (HLCross *HLCrossStrategy) Tick(records []Record) bool {
	//read config

	tradeAmount := Option["tradeAmount"]

	numTradeAmount, err := strconv.ParseFloat(Option["tradeAmount"], 64)
	if err != nil {
		logger.Errorln("config item tradeAmount is not float")
		return false
	}

	var Time []string
	var Price []float64
	var Volumn []float64
	for _, v := range records {
		Time = append(Time, v.TimeStr)
		Price = append(Price, v.Close)
		Volumn = append(Volumn, v.Volumn)
	}

	length := len(Price)

	if HLCross.PrevClosePrice != records[length-1].Close ||
		HLCross.PrevHighPrice != records[length-2].High ||
		HLCross.PrevLowPrice != records[length-2].Low {
		HLCross.PrevClosePrice = records[length-1].Close
		HLCross.PrevHighPrice = records[length-2].High
		HLCross.PrevLowPrice = records[length-2].Low

		logger.Infof("nowClose %0.02f prevHigh %0.02f prevLow %0.02f\n", records[length-1].Close, records[length-2].High, records[length-2].Low)
	}

	//HLCross cross
	if records[length-1].Close > records[length-2].High {
		if Option["enable_trading"] == "1" && PrevTrade != "buy" {
			if GetAvailable_coin() < numTradeAmount {
				warning = "HLCross up, but 没有足够的法币可买"
				PrevTrade = "buy"
			} else {
				warning = "HLCross up, 买入buy In<----市价" + getTradePrice("", Price[length-1]) +
					",委托价" + getTradePrice("buy", Price[length-1])
				logger.Infoln(warning)
				if Buy(getTradePrice("buy", Price[length-1]), tradeAmount) != "0" {
					PrevBuyPirce = Price[length-1]
					warning += "[委托成功]"
					PrevTrade = "buy"
				} else {
					warning += "[委托失败]"
				}
			}

			go email.TriggerTrender(warning)
		}
	} else if records[length-1].Close < records[length-2].Low {
		if Option["enable_trading"] == "1" && PrevTrade != "sell" {
			if GetAvailable_coin() < numTradeAmount {
				warning = "HLCross down, but 没有足够的币可卖"
				PrevTrade = "sell"
				PrevBuyPirce = 0
			} else {
				warning = "HLCross down, 卖出Sell Out---->市价" + getTradePrice("", Price[length-1]) +
					",委托价" + getTradePrice("sell", Price[length-1])
				if Sell(getTradePrice("sell", Price[length-1]), tradeAmount) != "0" {
					warning += "[委托成功]"
					PrevTrade = "sell"
					PrevBuyPirce = 0
				} else {
					warning += "[委托失败]"
				}
			}

			logger.Infoln(warning)

			go email.TriggerTrender(warning)
		}
	}

	//do sell when price is below stoploss point
	stop_loss_detect(Price)

	return true
}
