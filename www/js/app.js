var normalColour = 'rgb(178, 178, 178)',
    flashColour = '#FFF';

function formatPrice(f) {
  return f.toFixed(2);
}

function initPrices(prices) {
  var template = $('#template-price').html();

  for(var price in prices) {
    var item = template
                 .replace(/__NAME__/, price)
                 .replace(/__ID__/, 'price-' + price)
                 .replace(/__PRICE__/, formatPrice(prices[price]));
    $('.prices').append(item);
  }
}

function updatePrices(prices) {
  for(var price in prices) {
    var item = $('.price-' + price),
        value = formatPrice(prices[price]);

    if(item.text() == value) {
      continue;
    }

    item
      .text(value)
      .css('color', flashColour);

    setTimeout(function() {
      item.css('color', normalColour);
    }, 1000);
  }
}

function handlePrices(prices) {
  if($('.prices').children().length) {
    updatePrices(prices);
  } else {
    initPrices(prices);
  }
}

function initAssets(assets) {
  var templateVenue = $('#template-venue').html(),
      templateAsset = $('#template-asset').html();

  for(var venue in assets) {
    var $venue = $(templateVenue.replace(/__NAME__/, venue)),
        container = $venue.find('.venue-assets');

    for(var asset in assets[venue]) {
      var item = templateAsset
                   .replace(/__NAME__/, asset)
                   .replace(/__ID__/, 'asset-' + venue + '-' + asset)
                   .replace(/__QUANTITY__/, assets[venue][asset]);
      container.append(item);
    }

    $('.assets').append($venue);
  }
}

function updateAssets(assets) {
  for(var venue in assets) {
    for(var asset in assets[venue]) {
      var item = $('.asset-' +  + venue + '-' + asset),
          value = assets[venue][asset];

      if(item.text() == value) {
        continue;
      }

      item
        .text(value)
        .css('color', flashColour);

      setTimeout(function() {
        item.css('color', normalColour);
      }, 1000);
    }
  }
}

function handleAssets(assets) {
  if($('.assets').children().length) {
    updateAssets(assets);
  } else {
    initAssets(assets);
  }
}

function handlePayload(json) {
  if(json.assets) {
    handleAssets(json.assets);
  }

  if(json.prices) {
    handlePrices(json.prices);
  }

  var stats = ['exposure', 'cost', 'value', 'pnl', 'pnl_percentage'];
  for(var i in stats) {
    var stat = stats[i];

    if(!json.hasOwnProperty(stat)) {
      continue;
    }

    var value = json[stat];
    if(stat == 'exposure') {
      value = value.toFixed(8);
    } else if(['cost', 'value', 'pnl'].includes(stat)) {
      value = value.toLocaleString('en-US', { style: 'currency', currency: 'USD' });
    } else if(stat == 'pnl_percentage') {
      value = value.toFixed(2) + '%';
    }

    $('.stat-' + stat).text(value);
  }
}

$(function() {
  var ws = new WebSocket('ws://' + window.location.host + '/ws');

  ws.onopen = function() {
    console.log('> onopen');
    var token = readCookie('auth_token') || window.prompt('Auth token');
    createCookie('auth_token', token, 365);
    ws.send(JSON.stringify({auth: token}));
  }

  ws.onclose = function() {
    console.log('> onclose');
  }

  ws.onerror = function(error) {
    console.log('> onerror');
    console.log(error);
  }

  ws.onmessage = function(message) {
    var json = JSON.parse(message.data);

    if(json.error) {
      window.alert('Authentication failed');
      eraseCookie('auth_token');
      return;
    }

    handlePayload(json);
  }
});
