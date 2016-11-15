'use strict';

function initMap() {
	var sf = {lat: 37.749, lng: -122.439};
	var mapElement = document.getElementById('map');
	var map = new google.maps.Map(mapElement, {
		zoom: 11,
		center: sf
	});
	
	var $locations = $('.location');
	
	$locations.each(function () {
		var $location = $(this);
		var name = $location.data('name');
		var lat = $location.data('lat');
		var lng = $location.data('lng');
		
		console.log(lat, lng);
		
		if (lat && lng) {
			var marker = new google.maps.Marker({
				title: name,
				position: {lat: lat, lng: lng},
				animation: google.maps.Animation.DROP,
				map: map
			});
			
			$location.hover(function () {
				marker.setAnimation(google.maps.Animation.BOUNCE);
			}, function () {
				marker.setAnimation(null);
			})
		} else {
			$location.addClass('warning');
		}
	});
}
