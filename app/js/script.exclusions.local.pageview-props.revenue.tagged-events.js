!function(){"use strict";var o=window.location,l=window.document,s=l.currentScript,p=s.getAttribute("data-api")||new URL(s.src).origin+"/api/event";function c(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return c("localStorage flag",t)}catch(e){}var n=s&&s.getAttribute("data-include"),r=s&&s.getAttribute("data-exclude");if("pageview"===e){n=!n||n.split(",").some(a),r=r&&r.split(",").some(a);if(!n||r)return c("exclusion rule",t)}function a(e){return o.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var n={},r=(n.n=e,n.u=o.href,n.d=s.getAttribute("data-domain"),n.r=l.referrer||null,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props),t&&t.revenue&&(n.$=t.revenue),s.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),i=n.p||{},u=(r.forEach(function(e){var t=e.replace("event-",""),e=s.getAttribute(e);i[t]=i[t]||e}),n.p=i,new XMLHttpRequest);u.open("POST",p,!0),u.setRequestHeader("Content-Type","text/plain"),u.send(JSON.stringify(n)),u.onreadystatechange=function(){4===u.readyState&&t&&t.callback&&t.callback({status:u.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var n,r=0;r<t.length;r++)e.apply(this,t[r]);function a(){n!==o.pathname&&(n=o.pathname,e("pageview"))}var i,u=window.history;function f(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}u.pushState&&(i=u.pushState,u.pushState=function(){i.apply(this,arguments),a()},window.addEventListener("popstate",a)),"prerender"===l.visibilityState?l.addEventListener("visibilitychange",function(){n||"visible"!==l.visibilityState||a()}):a();var v=1;function d(e){"auxclick"===e.type&&e.button!==v||((e=function(e){for(;e&&(void 0===e.tagName||!f(e)||!e.href);)e=e.parentNode;return e}(e.target))&&e.href&&e.href.split("?")[0],function e(t,n){if(!t||b<n)return!1;if(w(t))return!0;return e(t.parentNode,n+1)}(e,0))}function m(e,t,n){var r,a=!1;function i(){a||(a=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((r={props:n.props}).revenue=n.revenue,plausible(n.name,r)):((r={props:n.props,callback:i}).revenue=n.revenue,plausible(n.name,r),setTimeout(i,5e3),e.preventDefault())}function g(e){var e=w(e)?e:e&&e.parentNode,t={name:null,props:{},revenue:{}},n=e&&e.classList;if(n)for(var r=0;r<n.length;r++){var a,i,u=n.item(r),o=u.match(/plausible-event-(.+)(=|--)(.+)/),o=(o&&(a=o[1],i=o[3].replace(/\+/g," "),"name"==a.toLowerCase()?t.name=i:t.props[a]=i),u.match(/plausible-revenue-(.+)(=|--)(.+)/));o&&(a=o[1],i=o[3],t.revenue[a]=i)}return t}l.addEventListener("click",d),l.addEventListener("auxclick",d);var b=3;function h(e){if("auxclick"!==e.type||e.button===v){for(var t,n,r,a,i=e.target,u=0;u<=b&&i;u++){if((r=i)&&r.tagName&&"form"===r.tagName.toLowerCase())return;f(i)&&(t=i),w(i)&&(n=i),i=i.parentNode}n&&(a=g(n),t?(a.props.url=t.href,m(e,t,a)):((e={}).props=a.props,e.revenue=a.revenue,plausible(a.name,e)))}}function w(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}l.addEventListener("submit",function(e){var t,n=e.target,r=g(n);function a(){t||(t=!0,n.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),(e={props:r.props,callback:a}).revenue=r.revenue,plausible(r.name,e))}),l.addEventListener("click",h),l.addEventListener("auxclick",h)}();